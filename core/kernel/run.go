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
	err = m.land.RangeList(func(tri Tripod) error {
		tri.StartBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		tri.EndBlock(newBlock)
		return nil
	})
	if err != nil {
		return err
	}

	// finalize this block
	return m.land.RangeList(func(tri Tripod) error {
		tri.FinalizeBlock(newBlock)
		return nil
	})
}

func (m *Kernel) makeNewBasicBlock() (*CompactBlock, error) {
	var newBlock *CompactBlock = m.chain.NewEmptyBlock()

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

func (m *Kernel) ExecuteTxns(block *CompactBlock) error {
	stxns, err := m.base.GetTxns(block.Hash)
	if err != nil {
		return err
	}

	var results []Result
	for _, stxn := range stxns {
		ecall := stxn.Raw.Ecall
		ctx, err := context.NewContext(stxn.Pubkey.Address(), ecall.Params)
		if err != nil {
			return err
		}

		exec, lei, err := m.land.GetExecLei(ecall)
		if err != nil {
			m.handleError(err, ctx, block, stxn)
			continue
		}

		if IfLeiOut(lei, block) {
			m.handleError(OutOfLei, ctx, block, stxn)
			break
		}

		err = exec(ctx, block)
		if err != nil {
			m.stateDB.Discard()
			m.handleError(err, ctx, block, stxn)
		} else {
			m.stateDB.NextTxn()
		}

		block.UseLei(lei)

		m.handleEvent(ctx, block, stxn)

		for _, e := range ctx.Events {
			results = append(results, e)
		}
		if ctx.Error != nil {
			results = append(results, ctx.Error)
		}

		err = m.base.SetEvents(ctx.Events)
		if err != nil {
			return err
		}
		err = m.base.SetError(ctx.Error)
		if err != nil {
			return err
		}
	}

	stateRoot, err := m.stateDB.Commit()
	if err != nil {
		return err
	}
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

func (m *Kernel) handleError(err error, ctx *context.Context, block *CompactBlock, stxn *SignedTxn) {
	ctx.EmitError(err)
	ecall := stxn.Raw.Ecall

	ctx.Error.Caller = stxn.Raw.Caller
	ctx.Error.BlockStage = ExecuteTxnsStage
	ctx.Error.TripodName = ecall.TripodName
	ctx.Error.ExecName = ecall.ExecName
	ctx.Error.BlockHash = block.Hash
	ctx.Error.Height = block.Height

	logrus.Error("push error: ", ctx.Error.Error())
	if m.sub != nil {
		m.sub.Emit(ctx.Error)
	}

}

func (m *Kernel) handleEvent(ctx *context.Context, block *CompactBlock, stxn *SignedTxn) {
	for _, event := range ctx.Events {
		ecall := stxn.Raw.Ecall

		event.Height = block.Height
		event.BlockHash = block.Hash
		event.ExecName = ecall.ExecName
		event.TripodName = ecall.TripodName
		event.BlockStage = ExecuteTxnsStage
		event.Caller = stxn.Raw.Caller

		if m.sub != nil {
			m.sub.Emit(event)
		}
	}
}
