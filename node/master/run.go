package master

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/context"
	. "github.com/yu-org/yu/node"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/txn"
	ytime "github.com/yu-org/yu/utils/time"
	. "github.com/yu-org/yu/yerror"
)

func (m *Master) Run() {

	switch m.RunMode {
	case LocalNode:
		for {
			err := m.LocalRun()
			if err != nil {
				logrus.Errorf("local-run blockchain error: %s", err.Error())
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

func (m *Master) LocalRun() (err error) {
	newBlock, err := m.makeNewBasicBlock()
	if err != nil {
		return err
	}

	// start a new block
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.StartBlock(newBlock)
	})
	if err != nil {
		return err
	}

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(newBlock)
	})
	if err != nil {
		return err
	}

	// finalize this block
	return m.land.RangeList(func(tri Tripod) error {
		return tri.FinalizeBlock(newBlock)
	})
}

func (m *Master) makeNewBasicBlock() (IBlock, error) {
	var newBlock IBlock = m.chain.NewEmptyBlock()

	newBlock.SetTimestamp(ytime.NowNanoTsU64())
	prevBlock, err := m.chain.GetEndBlock()
	if err != nil {
		return nil, err
	}
	newBlock.SetPreHash(prevBlock.GetHash())
	newBlock.SetPeerID(m.host.ID())
	newBlock.SetHeight(prevBlock.GetHeight() + 1)
	newBlock.SetLeiLimit(m.leiLimit)
	return newBlock, nil
}

func (m *Master) ExecuteTxns(block IBlock) error {
	stxns, err := m.base.GetTxns(block.GetHash())
	if err != nil {
		return err
	}
	for _, stxn := range stxns {
		ecall := stxn.GetRaw().GetEcall()
		ctx, err := context.NewContext(stxn.GetPubkey().Address(), ecall.Params)
		if err != nil {
			return err
		}

		exec, lei, err := m.land.GetExecLei(ecall)
		if err != nil {
			m.handleError(err, ctx, block, stxn)
			continue
		}

		if IfLeiOut(lei, block) {
			m.handleError(OutOfEnergy, ctx, block, stxn)
			break
		}

		err = exec(ctx, block)
		if err != nil {
			m.stateStore.Discard()
			m.handleError(err, ctx, block, stxn)
		} else {
			m.stateStore.NextTxn()
		}

		block.UseLei(lei)

		m.handleEvent(ctx, block, stxn)

		err = m.base.SetEvents(ctx.Events)
		if err != nil {
			return err
		}
		err = m.base.SetError(ctx.Error)
		if err != nil {
			return err
		}
	}

	stateRoot, err := m.stateStore.Commit()
	if err != nil {
		return err
	}
	block.SetStateRoot(stateRoot)

	return nil
}

func (m *Master) MasterWokrerRun() error {
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

func (m *Master) nortifyWorker(workersIps []string, path string, newBlock IBlock) error {
	blockByt, err := newBlock.Encode()
	if err != nil {
		return err
	}

	for _, ip := range workersIps {
		resp, err := PostRequest(ip+path, blockByt)
		if err != nil {
			return err
		}
		respBlock, err := DecodeBlockFromHttp(resp.Body, m.chain)
		if err != nil {
			return err
		}
		newBlock = respBlock
	}
	return nil
}

func (m *Master) handleError(err error, ctx *context.Context, block IBlock, stxn *txn.SignedTxn) {
	ctx.EmitError(err)
	ecall := stxn.GetRaw().GetEcall()

	ctx.Error.Caller = stxn.GetRaw().GetCaller()
	ctx.Error.BlockStage = ExecuteTxnsStage
	ctx.Error.TripodName = ecall.TripodName
	ctx.Error.ExecName = ecall.ExecName
	ctx.Error.BlockHash = block.GetHash()
	ctx.Error.Height = block.GetHeight()

	logrus.Error("push error: ", ctx.Error.Error())
	if m.sub != nil {
		m.sub.Push(ctx.Error)
	}

}

func (m *Master) handleEvent(ctx *context.Context, block IBlock, stxn *txn.SignedTxn) {
	for _, event := range ctx.Events {
		ecall := stxn.GetRaw().GetEcall()

		event.Height = block.GetHeight()
		event.BlockHash = block.GetHash()
		event.ExecName = ecall.ExecName
		event.TripodName = ecall.TripodName
		event.BlockStage = ExecuteTxnsStage
		event.Caller = stxn.GetRaw().GetCaller()

		if m.sub != nil {
			m.sub.Push(event)
		}
	}
}
