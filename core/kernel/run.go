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
			err := k.LocalRun()
			if err != nil {
				logrus.Panicf("local-run blockchain error: %s", err.Error())
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

	// end block and append to chain
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
	newBlock := k.chain.NewEmptyBlock()

	newBlock.Timestamp = ytime.NowTsU64()
	prevBlock, err := k.chain.GetEndBlock()
	if err != nil {
		return nil, err
	}
	newBlock.PrevHash = prevBlock.Hash
	newBlock.PeerID = k.p2pNetwork.LocalID()
	newBlock.Height = prevBlock.Height + 1
	newBlock.LeiLimit = k.leiLimit
	return newBlock, nil
}

func (k *Kernel) OrderedExecute(block *Block) error {
	stxns := block.Txns

	var results []*Result

	for _, stxn := range stxns {
		wrCall := stxn.Raw.WrCall
		ctx, err := context.NewWriteContext(stxn, block)
		if err != nil {
			return err
		}

		writing, err := k.land.GetWriting(wrCall.TripodName, wrCall.FuncName)
		if err != nil {
			k.handleError(err, ctx, block, stxn)
			continue
		}

		err = writing(ctx)
		if IfLeiOut(ctx.LeiCost, block) {
			k.stateDB.Discard()
			k.handleError(OutOfLei, ctx, block, stxn)
			break
		}
		if err != nil {
			k.stateDB.Discard()
			k.handleError(err, ctx, block, stxn)
		} else {
			k.stateDB.NextTxn()
		}

		block.UseLei(ctx.LeiCost)

		// if no error and event, give a default event
		if ctx.Error == nil && len(ctx.Events) == 0 {
			_ = ctx.EmitJsonEvent(DefaultJsonEvent)
		}

		k.handleEvent(ctx, block, stxn)

		for _, e := range ctx.Events {
			results = append(results, NewEvent(e))
		}
		if ctx.Error != nil {
			results = append(results, NewError(ctx.Error))
		}
	}

	if len(results) > 0 {
		err := k.txDB.SetResults(results)
		if err != nil {
			return err
		}
	}

	stateRoot, err := k.stateDB.Commit()
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
	//newBlock := k.chain.NewDefaultBlock()
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

func (k *Kernel) handleError(err error, ctx *context.WriteContext, block *Block, stxn *SignedTxn) {
	ctx.EmitError(err)
	wrCall := stxn.Raw.WrCall

	ctx.Error.Caller = stxn.GetCallerAddr()
	ctx.Error.BlockStage = ExecuteTxnsStage
	ctx.Error.TripodName = wrCall.TripodName
	ctx.Error.WritingName = wrCall.FuncName
	ctx.Error.BlockHash = block.Hash
	ctx.Error.Height = block.Height

	logrus.Error("push error: ", ctx.Error.Error())
	if k.sub != nil {
		k.sub.Emit(NewError(ctx.Error))
	}

}

func (k *Kernel) handleEvent(ctx *context.WriteContext, block *Block, stxn *SignedTxn) {
	for _, event := range ctx.Events {
		wrCall := stxn.Raw.WrCall

		event.Height = block.Height
		event.BlockHash = block.Hash
		event.WritingName = wrCall.FuncName
		event.TripodName = wrCall.TripodName
		event.LeiCost = ctx.LeiCost
		event.BlockStage = ExecuteTxnsStage
		event.Caller = stxn.GetCallerAddr()

		if k.sub != nil {
			k.sub.Emit(NewEvent(event))
		}
	}
}
