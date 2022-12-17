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

func (m *Kernel) Run() {

	switch m.RunMode {
	case LocalNode:
		for {
			err := m.LocalRun()
			if err != nil {
				logrus.Panicf("local-run blockchain error: %s", err.Error())
			}
		}
	case MasterWorker:
		for {
			err := m.MasterWokrerRun()
			logrus.Errorf("master-worker-run blockchain error: %s", err.Error())
		}

	default:
		logrus.Panic(NoRunMode)
	}

}

func (m *Kernel) LocalRun() (err error) {
	newBlock, err := m.makeNewBasicBlock()
	if err != nil {
		return err
	}

	// start a new block
	err = m.land.RangeList(func(tri *Tripod) error {
		tri.StartBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// end block and append to chain
	err = m.land.RangeList(func(tri *Tripod) error {
		tri.EndBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// finalize this block
	return m.land.RangeList(func(tri *Tripod) error {
		tri.FinalizeBlock(newBlock)
		return nil
	})
}

func (m *Kernel) makeNewBasicBlock() (*Block, error) {
	newBlock := m.chain.NewEmptyBlock()

	newBlock.Timestamp = ytime.NowNanoTsU64()
	prevBlock, err := m.chain.GetEndBlock()
	if err != nil {
		return nil, err
	}
	newBlock.PrevHash = prevBlock.Hash
	newBlock.PeerID = m.p2pNetwork.LocalID()
	newBlock.Height = prevBlock.Height + 1
	newBlock.LeiLimit = m.leiLimit
	return newBlock, nil
}

func (m *Kernel) OrderedExecute(block *Block) error {
	stxns := block.Txns

	var results []Result

	for _, stxn := range stxns {
		wrCall := stxn.Raw.WrCall
		ctx, err := context.NewWriteContext(stxn, block)
		if err != nil {
			return err
		}

		writing, err := m.land.GetWriting(wrCall)
		if err != nil {
			m.handleError(err, ctx, block, stxn)
			continue
		}

		err = writing(ctx)
		if IfLeiOut(ctx.LeiCost, block) {
			m.stateDB.Discard()
			m.handleError(OutOfLei, ctx, block, stxn)
			break
		}
		if err != nil {
			m.stateDB.Discard()
			m.handleError(err, ctx, block, stxn)
		} else {
			m.stateDB.NextTxn()
		}

		block.UseLei(ctx.LeiCost)

		m.handleEvent(ctx, block, stxn)

		for _, e := range ctx.Events {
			results = append(results, e)
		}
		if ctx.Error != nil {
			results = append(results, ctx.Error)
		}
	}

	// resStart := time.Now()
	if len(results) > 0 {
		err := m.base.SetResults(results)
		if err != nil {
			return err
		}
	}
	// logrus.Infof("----- setResults costs %d ms", time.Since(resStart).Milliseconds())

	// sStart := time.Now()
	stateRoot, err := m.stateDB.Commit()
	if err != nil {
		return err
	}
	// logrus.Infof("---!!!--- stateDB costs %d ms", time.Since(sStart).Milliseconds())

	block.StateRoot = stateRoot

	block.ReceiptRoot, err = CaculateReceiptRoot(results)
	return err
}

func (m *Kernel) MasterWokrerRun() error {
	//workersIps, err := m.allWorkersIP()
	//if err != nil {
	//	return err
	//}
	//
	//newBlock := m.chain.NewDefaultBlock()
	//
	//err = m.nortifyWorker(workersIps, StartBlockPath, newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//// todo: if need broadcast block,
	//// m.readyBroadcastBlock(newBlock)
	//
	//err = m.SyncTxns(newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//err = m.nortifyWorker(workersIps, EndBlockPath, newBlock)
	//if err != nil {
	//	return err
	//}
	//
	//go func() {
	//	err := m.nortifyWorker(workersIps, ExecuteTxnsPath, newBlock)
	//	if err != nil {
	//		logrus.Errorf("nortify worker executing txns error: %s", err.Error())
	//	}
	//}()
	//
	//return m.nortifyWorker(workersIps, FinalizeBlockPath, newBlock)
	return nil
}

func (m *Kernel) handleError(err error, ctx *context.WriteContext, block *Block, stxn *SignedTxn) {
	ctx.EmitError(err)
	wrCall := stxn.Raw.WrCall

	ctx.Error.Caller = stxn.Raw.Caller
	ctx.Error.BlockStage = ExecuteTxnsStage
	ctx.Error.TripodName = wrCall.TripodName
	ctx.Error.WritingName = wrCall.WritingName
	ctx.Error.BlockHash = block.Hash
	ctx.Error.Height = block.Height

	logrus.Error("push error: ", ctx.Error.Error())
	if m.sub != nil {
		m.sub.Emit(ctx.Error)
	}

}

func (m *Kernel) handleEvent(ctx *context.WriteContext, block *Block, stxn *SignedTxn) {
	for _, event := range ctx.Events {
		wrCall := stxn.Raw.WrCall

		event.Height = block.Height
		event.BlockHash = block.Hash
		event.WritingName = wrCall.WritingName
		event.TripodName = wrCall.TripodName
		event.LeiCost = ctx.LeiCost
		event.BlockStage = ExecuteTxnsStage
		event.Caller = stxn.Raw.Caller

		if m.sub != nil {
			m.sub.Emit(event)
		}
	}
}
