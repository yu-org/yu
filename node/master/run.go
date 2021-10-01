package master

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/node"
	. "github.com/yu-org/yu/tripod"
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
		return tri.StartBlock(newBlock, m.GetEnv(), m.land)
	})
	if err != nil {
		return err
	}

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(newBlock, m.GetEnv(), m.land)
	})
	if err != nil {
		return err
	}

	// finalize this block
	return m.land.RangeList(func(tri Tripod) error {
		return tri.FinalizeBlock(newBlock, m.GetEnv(), m.land)
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

//func (m *Master) broadcastBlockAndTxns(b IBlock) error {
//	err := m.pubBlock(b)
//	if err != nil {
//		return err
//	}
//
//	blockHash := b.GetHash()
//	txns, err := m.base.GetTxns(blockHash)
//	if err != nil {
//		return err
//	}
//
//	if len(txns) == 0 {
//		return nil
//	}
//
//	logrus.Warnf("=== pub block(%s) to P2P ===", blockHash.String())
//	for _, stxn := range txns {
//		logrus.Warnf("============== pub stxn(%s) to P2P ============", stxn.TxnHash.String())
//	}
//
//	return m.pubPackedTxns(blockHash, txns)
//}
