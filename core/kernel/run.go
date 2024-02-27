package kernel

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
	ytime "github.com/yu-org/yu/utils/time"
)

var DefaultJsonEvent = map[string]string{"status": "ok"}

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
				err := k.LocalRun()
				if err != nil {
					logrus.Panicf("local-run blockchain error: %s", err.Error())
				}
			}

		}
	case MasterWorker:
		for {
			err := k.MasterWokrerRun()
			logrus.Errorf("master-worker-run blockchain error: %s", err.Error())
		}

	default:
		logrus.Panic(NoRunMode)
	}

}

func (k *Kernel) LocalRun() (err error) {
	newBlock, err := k.makeNewBasicBlock()
	if err != nil {
		return err
	}

	// start a new block
	err = k.land.RangeList(func(tri *Tripod) error {
		tri.StartBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// end block and append to Chain
	err = k.land.RangeList(func(tri *Tripod) error {
		tri.EndBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// finalize this block
	return k.land.RangeList(func(tri *Tripod) error {
		tri.FinalizeBlock(newBlock)
		return nil
	})
}

func (k *Kernel) makeNewBasicBlock() (*Block, error) {
	newBlock := k.Chain.NewEmptyBlock()

	newBlock.Timestamp = ytime.NowTsU64()
	prevBlock, err := k.Chain.GetEndBlock()
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

	var results []*Receipt

	for _, stxn := range stxns {
		wrCall := stxn.Raw.WrCall
		ctx, err := context.NewWriteContext(stxn, block)
		if err != nil {
			result := k.handleError(err, ctx, block, stxn)
			results = append(results, result)
			continue
		}

		writing, _ := k.land.GetWriting(wrCall.TripodName, wrCall.FuncName)

		err = writing(ctx)
		if IfLeiOut(ctx.LeiCost, block) {
			k.State.Discard()
			result := k.handleError(OutOfLei, ctx, block, stxn)
			results = append(results, result)
			break
		}
		if err != nil {
			k.State.Discard()
			k.handleError(err, ctx, block, stxn)
		} else {
			k.State.NextTxn()
		}

		block.UseLei(ctx.LeiCost)

		// if no error and event, give a default event
		//if ctx.Error == nil && len(ctx.Events) == 0 {
		//	_ = ctx.EmitJsonEvent(DefaultJsonEvent)
		//}

		result := k.handleEvent(ctx, block, stxn)

		results = append(results, result)
	}

	if len(results) > 0 {
		err := k.TxDB.SetResults(results)
		if err != nil {
			return err
		}
	}

	stateRoot, err := k.State.Commit()
	if err != nil {
		return err
	}

	block.StateRoot = BytesToHash(stateRoot)

	block.ReceiptRoot, err = CaculateReceiptRoot(results)
	return err
}

func (k *Kernel) MasterWokrerRun() error {
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

func (k *Kernel) handleError(err error, ctx *context.WriteContext, block *Block, stxn *SignedTxn) *Receipt {
	logrus.Error("push error: ", err.Error())
	ctx.EmitError(err)
	result := NewResult(ctx.Events, err)
	k.handleResult(ctx, result, block, stxn)
	return result
}

func (k *Kernel) handleEvent(ctx *context.WriteContext, block *Block, stxn *SignedTxn) *Receipt {
	result := NewWithEvents(ctx.Events)
	k.handleResult(ctx, result, block, stxn)
	return result
}

func (k *Kernel) handleResult(ctx *context.WriteContext, result *Receipt, block *Block, stxn *SignedTxn) {
	wrCall := stxn.Raw.WrCall

	result.Caller = stxn.GetCallerAddr()
	result.BlockStage = ExecuteTxnsStage
	result.TripodName = wrCall.TripodName
	result.WritingName = wrCall.FuncName
	result.BlockHash = block.Hash
	result.Height = block.Height
	result.LeiCost = ctx.LeiCost

	if k.Sub != nil {
		k.Sub.Emit(result)
	}
}
