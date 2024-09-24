package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
	ytime "github.com/yu-org/yu/utils/time"
)

func (k *Kernel) Run() {
	go func() {
		for {
			err := k.AcceptUnpkgTxns()
			if err != nil {
				logrus.Errorf("accept unpacked txns error: %s", err.Error())
			}
		}

	}()

	switch k.RunMode {
	case LocalNode:
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
	case MasterWorker:
		for {
			err := k.MasterWorkerRun()
			logrus.Errorf("master-worker-run blockchain error: %s", err.Error())
		}

	default:
		logrus.Panic(NoRunMode)
	}

}

func (k *Kernel) LocalRun() (newBlock *Block, err error) {
	newBlock, err = k.makeNewBasicBlock()
	if err != nil {
		return
	}

	// start a new block
	err = k.Land.RangeList(func(tri *Tripod) error {
		// start := time.Now()
		tri.BlockCycle.StartBlock(newBlock)
		// metrics.StartBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	if err != nil {
		return
	}

	// end block
	err = k.Land.RangeList(func(tri *Tripod) error {
		// start := time.Now()
		tri.BlockCycle.EndBlock(newBlock)
		// metrics.EndBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	if err != nil {
		return
	}

	// finalize this block
	err = k.Land.RangeList(func(tri *Tripod) error {
		// start := time.Now()
		tri.BlockCycle.FinalizeBlock(newBlock)
		// metrics.FinalizeBlockDuration.WithLabelValues(strconv.FormatInt(int64(newBlock.Height), 10), tri.Name()).Observe(time.Now().Sub(start).Seconds())
		return nil
	})
	return
}

func (k *Kernel) makeGenesisBlock() *Block {
	genesisBlock := k.Chain.NewEmptyBlock()

	genesisBlock.Timestamp = ytime.NowTsU64()
	genesisBlock.PeerID = k.P2pNetwork.LocalID()
	genesisBlock.Height = 0
	genesisBlock.LeiLimit = k.leiLimit
	return genesisBlock
}

func (k *Kernel) makeNewBasicBlock() (*Block, error) {
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

func (k *Kernel) OrderedExecute(block *Block) error {
	stxns := block.Txns

	receipts := make(map[Hash]*Receipt)

	for i, stxn := range stxns {
		wrCall := stxn.Raw.WrCall
		ctx, err := context.NewWriteContext(stxn, block, i)
		if err != nil {
			receipt := k.HandleError(err, ctx, block, stxn)
			receipts[stxn.TxnHash] = receipt
			continue
		}

		writing, _ := k.Land.GetWriting(wrCall.TripodName, wrCall.FuncName)

		err = writing(ctx)
		if IfLeiOut(ctx.LeiCost, block) {
			k.State.Discard()
			receipt := k.HandleError(OutOfLei, ctx, block, stxn)
			receipts[stxn.TxnHash] = receipt
			break
		}
		if err != nil {
			k.State.Discard()
			k.HandleError(err, ctx, block, stxn)
		} else {
			k.State.NextTxn()
		}

		block.UseLei(ctx.LeiCost)

		// if no error and event, give a default event
		//if ctx.Error == nil && len(ctx.Events) == 0 {
		//	_ = ctx.EmitJsonEvent(DefaultJsonEvent)
		//}

		receipt := k.HandleEvent(ctx, block, stxn)
		receipts[stxn.TxnHash] = receipt
	}
	return k.PostExecute(block, receipts)
}

func (k *Kernel) PostExecute(block *Block, receipts map[Hash]*Receipt) error {
	k.Land.RangeList(func(t *Tripod) error {
		t.Committer.Commit(block)
		return nil
	})

	if len(receipts) > 0 {
		err := k.TxDB.SetReceipts(receipts)
		if err != nil {
			return err
		}
	}

	stateRoot, err := k.State.Commit()
	if err != nil {
		return err
	}

	// Because tripod.Committer could update this field.
	if block.StateRoot == NullHash && stateRoot != nil {
		block.StateRoot = BytesToHash(stateRoot)
	}

	// Because tripod.Committer could update this field.
	if block.ReceiptRoot == NullHash {
		block.ReceiptRoot, err = CaculateReceiptRoot(receipts)
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

func (k *Kernel) HandleError(err error, ctx *context.WriteContext, block *Block, stxn *SignedTxn) *Receipt {
	logrus.Error("push error: ", err.Error())
	receipt := NewReceipt(ctx.Events, err, ctx.Extra)
	k.HandleReceipt(ctx, receipt, block, stxn)
	return receipt
}

func (k *Kernel) HandleEvent(ctx *context.WriteContext, block *Block, stxn *SignedTxn) *Receipt {
	receipt := NewReceipt(ctx.Events, nil, ctx.Extra)
	k.HandleReceipt(ctx, receipt, block, stxn)
	return receipt
}

func (k *Kernel) HandleReceipt(ctx *context.WriteContext, receipt *Receipt, block *Block, stxn *SignedTxn) {
	receipt.FillMetadata(block, stxn, ctx.LeiCost)
	receipt.BlockStage = ExecuteTxnsStage

	if k.Sub != nil {
		k.Sub.Emit(receipt)
	}
}
