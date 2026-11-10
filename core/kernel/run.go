package kernel

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/types"
	ytime "github.com/yu-org/yu/utils/time"
)

func (k *Kernel) AcceptUnpkgTxnsJob() {
	for {
		err := k.AcceptUnpackedTxns()
		if err != nil {
			logrus.Errorf("accept unpacked txns error: %s", err.Error())
		}
	}
}

func (k *Kernel) Run() {
	defer logrus.Info("Run exit")
	switch k.RunMode {
	case common.LocalNode:
		for {
			select {
			case <-k.stopChan:
				logrus.Info("Stop the Chain!")
				return
			default:
				block, err := k.LocalRun()
				if err != nil {
					logrus.Panicf("local-run blockchain error: %s on Block(%d)", err.Error(), block.Height)
				}
				if block.Height == k.cfg.MaxBlockNum {
					logrus.Infof("Stop the Chain on Block(%d)", block.Height)
					return
				}
			}

		}
	case common.MasterWorker:
		for {
			select {
			case <-k.stopChan:
				return
			default:
			}
			if err := k.MasterWorkerRun(); err != nil {
				logrus.Errorf("master-worker-run blockchain error: %s", err.Error())
			}
		}
	default:
		logrus.Panic(yerror.NoRunMode)
	}
}

func (k *Kernel) LocalRun() (newBlock *types.Block, err error) {
	newBlock, err = k.makeNewBasicBlock()
	if err != nil {
		return
	}

	// start a new block
	err = k.Land.RangeList(func(tri *tripod.Tripod) error {
		// start := time.Now()
		tri.BlockCycle.StartBlock(newBlock)
		// metrics.StartBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	if err != nil {
		return
	}

	// end block
	err = k.Land.RangeList(func(tri *tripod.Tripod) error {
		// start := time.Now()
		tri.BlockCycle.EndBlock(newBlock)
		// metrics.EndBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	if err != nil {
		return
	}

	// finalize this block
	err = k.Land.RangeList(func(tri *tripod.Tripod) error {
		// start := time.Now()
		tri.BlockCycle.FinalizeBlock(newBlock)
		// metrics.FinalizeBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	return
}

func (k *Kernel) makeNewBasicBlock() (*types.Block, error) {
	newBlock := k.Chain.NewEmptyBlock()

	newBlock.Timestamp = ytime.NowTsU64()
	prevBlock, err := k.Chain.GetEndCompactBlock()
	if err != nil {
		return nil, err
	}
	newBlock.PrevHash = prevBlock.Hash
	newBlock.PeerID = k.P2pNetwork.LocalID()
	newBlock.Height = prevBlock.Height + 1
	newBlock.LeiLimit = k.leiLimit
	return newBlock, nil
}

// SeqExecuteWritings executes a block against its parent state and rolls back failed transactions.
func (k *Kernel) SeqExecuteWritings(block *types.Block) (resultErr error) {
	k.executionMu.Lock()
	defer k.executionMu.Unlock()
	managed, transactional := k.State.(state.ExecutionState)
	if transactional {
		if err := managed.BeginExecution(block); err != nil {
			return err
		}
		k.receiptMu.Lock()
		k.pendingReceipts = nil
		k.receiptMu.Unlock()
		defer func() {
			if resultErr != nil {
				managed.AbortExecution()
			}
		}()
	}
	received := transactional && block.StateRoot != common.NullHash
	expectedLei := block.LeiUsed
	if transactional {
		block.LeiUsed = 0
	}
	stxns := block.Txns

	receipts := make(map[common.Hash]*types.Receipt)

	for i, stxn := range stxns {
		wrCall := stxn.Raw.WrCall
		ctx, err := context.NewWriteContext(stxn, block, i)
		if err != nil {
			return fmt.Errorf("invalid transaction context: %w", err)
		}
		write, err := k.Land.GetWriting(wrCall.TripodName, wrCall.FuncName)
		if err != nil {
			return err
		}

		err = write(ctx)
		if err == nil && transactional {
			err = managed.ExecutionError()
		}
		if types.IfLeiOut(ctx.LeiCost, block) {
			k.State.Discard()
			receipt := k.HandleError(yerror.OutOfLei, ctx, block, stxn)
			receipts[stxn.TxnHash] = receipt
			break
		}
		var receipt *types.Receipt
		if err != nil {
			k.State.Discard()
			receipt = k.HandleError(err, ctx, block, stxn)
		} else {
			k.State.NextTxn()
			receipt = k.HandleEvent(ctx, block, stxn)
		}

		block.UseLei(ctx.LeiCost)

		// if no error and event, give a default event
		//if ctx.Error == nil && len(ctx.Events) == 0 {
		//	_ = ctx.EmitJsonEvent(DefaultJsonEvent)
		//}

		receipts[stxn.TxnHash] = receipt
	}
	if received && block.LeiUsed != expectedLei {
		return yerror.OutOfLei
	}
	return k.PostExecute(block, receipts)
}

// PostExecute computes execution commitments before any SQL state version is published.
func (k *Kernel) PostExecute(block *types.Block, receipts map[common.Hash]*types.Receipt) error {
	k.Land.RangeList(func(t *tripod.Tripod) error {
		t.Committer.Commit(block)
		return nil
	})

	if _, ok := k.State.(state.ExecutionState); ok {
		calculated, err := types.CalculateExecutionReceiptRoot(receipts)
		if err != nil {
			return err
		}
		if block.ReceiptRoot != common.NullHash && block.ReceiptRoot != calculated {
			return fmt.Errorf("receipt root mismatch")
		}
		block.ReceiptRoot = calculated
	}
	_, managed := k.State.(state.ExecutionState)
	if len(receipts) > 0 && !managed {
		err := k.TxDB.SetReceipts(receipts)
		if err != nil {
			return err
		}
	}

	stateRoot, err := k.State.Commit()
	if err != nil {
		return err
	}

	if managed {
		k.receiptMu.Lock()
		if k.pendingReceipts == nil {
			k.pendingReceipts = make(map[*types.Block]map[common.Hash]*types.Receipt)
		}
		k.pendingReceipts[block] = receipts
		k.receiptMu.Unlock()
	}
	// Because tripod.Committer could update this field.
	if block.StateRoot == common.NullHash && stateRoot != nil {
		block.StateRoot = common.BytesToHash(stateRoot)
	}

	// Because tripod.Committer could update this field.
	if block.ReceiptRoot == common.NullHash {
		block.ReceiptRoot, err = types.CaculateReceiptRoot(receipts)
	}
	return err
}

func (k *Kernel) MasterWorkerRun() error {
	//workersIps, err := k.allWorkersIP()
	//if err != nil {
	//	return err
	//}
	//
	//newBlock := k.Chain.NewDefaultBlock()
	//
	//err = k.nortifyWorker(workersIps, StartBlockPath, newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//// todo: if need broadcast block,
	//// k.readyBroadcastBlock(newBlock)
	//
	//err = k.SyncTxns(newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//err = k.nortifyWorker(workersIps, EndBlockPath, newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//go func() {
	//	err := k.nortifyWorker(workersIps, ExecuteTxnsPath, newBlock)
	//	if err != nil {
	//		logrus.Errorf("nortify worker executing txns error: %s", err.Error())
	//	}
	//}()
	//
	//return k.nortifyWorker(workersIps, FinalizeBlockPath, newBlock)
	return nil
}

func (k *Kernel) HandleError(err error, ctx *context.WriteContext, block *types.Block, stxn *types.SignedTxn) *types.Receipt {
	logrus.Error("push error: ", err.Error())
	receipt := types.NewReceipt(ctx.Events, err, ctx.Extra)
	k.HandleReceipt(ctx, receipt, block, stxn)
	return receipt
}

func (k *Kernel) HandleEvent(ctx *context.WriteContext, block *types.Block, stxn *types.SignedTxn) *types.Receipt {
	receipt := types.NewReceipt(ctx.Events, nil, ctx.Extra)
	k.HandleReceipt(ctx, receipt, block, stxn)
	return receipt
}

func (k *Kernel) HandleReceipt(ctx *context.WriteContext, receipt *types.Receipt, block *types.Block, stxn *types.SignedTxn) {
	receipt.FillMetadata(block, stxn, ctx.LeiCost)
	receipt.BlockStage = common.ExecuteTxnsStage

	if _, managed := k.State.(state.ExecutionState); k.Sub != nil && !managed {
		k.Sub.Emit(receipt)
	}
}

// BindState binds a sealed execution and writes receipts with its final block hash.
// Consensus must call this before append; execution and binding must be serialized.
func (k *Kernel) BindState(block *types.Block) error {
	k.executionMu.Lock()
	defer k.executionMu.Unlock()
	managed, ok := k.State.(state.ExecutionState)
	if !ok {
		return nil
	}
	if err := managed.BindBlock(block); err != nil {
		return err
	}
	k.receiptMu.Lock()
	receipts, exists := k.pendingReceipts[block]
	delete(k.pendingReceipts, block)
	k.receiptMu.Unlock()
	if !exists {
		return fmt.Errorf("missing committed execution receipts")
	}
	for _, receipt := range receipts {
		receipt.BlockHash = block.Hash
	}
	if len(receipts) != 0 {
		if err := k.TxDB.SetReceipts(receipts); err != nil {
			return err
		}
	}
	if k.Sub != nil {
		for _, receipt := range receipts {
			k.Sub.Emit(receipt)
		}
	}
	return nil
}
