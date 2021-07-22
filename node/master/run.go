package master

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/tripod"
	ytime "github.com/Lawliet-Chan/yu/utils/time"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
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

func (m *Master) LocalRun() error {

	needBcBlock := false

	newBlock, err := m.makeNewBasicBlock()
	if err != nil {
		return err
	}

	// start a new block
	err = m.land.RangeList(func(tri Tripod) error {
		var (
			need bool
			err  error
		)
		need, err = tri.StartBlock(newBlock, m.GetEnv(), m.land)
		if err != nil {
			return err
		}
		if need {
			needBcBlock = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	if needBcBlock {
		go func() {
			err := m.pubBlock(newBlock)
			if err != nil {
				logrus.Errorf("broadcast block(%s) and txns error: %s", newBlock.GetHash().String(), err.Error())
			}
		}()
	} else {
		err = m.SyncTxns(newBlock)
		if err != nil {
			return err
		}
		err = m.txPool.RemoveTxns(newBlock.GetTxnsHashes())
		if err != nil {
			return err
		}
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
	newBlock.SetHeight(prevBlock.GetHeight() + 1)
	newBlock.SetEnergyLimit(m.energyLimit)
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
