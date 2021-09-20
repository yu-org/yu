package master

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/common"
	. "github.com/yu-altar/yu/node"
	. "github.com/yu-altar/yu/tripod"
	ytime "github.com/yu-altar/yu/utils/time"
	. "github.com/yu-altar/yu/yerror"
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

	m.subMsgs()

	var broadcastMsg []byte

	newBlock, err := m.makeNewBasicBlock()
	if err != nil {
		return err
	}

	// start a new block
	err = m.land.RangeList(func(tri Tripod) error {
		msgByt, err := tri.StartBlock(newBlock, m.GetEnv(), m.land, m.msgOnStart)
		if err != nil {
			return err
		}

		if msgByt != nil {
			broadcastMsg = msgByt
		}
		return nil
	})
	if err != nil {
		return err
	}

	if broadcastMsg != nil {
		go func(msg []byte) {
			err := m.pubToP2P(m.startBlockTopic, msg)
			if err != nil {
				logrus.Errorf("broadcast message on Start Block error: %s", newBlock.GetHash().String(), err.Error())
			}
		}(broadcastMsg)
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
		broadcastMsg, err = tri.EndBlock(newBlock, m.GetEnv(), m.land, m.msgOnEnd)
		return err
	})
	if err != nil {
		return err
	}

	if broadcastMsg != nil {
		go func(msg []byte) {
			err := m.pubToP2P(m.endBlockTopic, msg)
			if err != nil {
				logrus.Errorf("broadcast message on End Block error: %s", newBlock.GetHash().String(), err.Error())
			}
		}(broadcastMsg)
	}

	// finalize this block
	err = m.land.RangeList(func(tri Tripod) error {
		broadcastMsg, err = tri.FinalizeBlock(newBlock, m.GetEnv(), m.land, m.msgOnFinalize)
		return err
	})
	if err != nil {
		return err
	}

	if broadcastMsg != nil {
		go func(msg []byte) {
			err := m.pubToP2P(m.finalizeBlockTopic, msg)
			if err != nil {
				logrus.Errorf("broadcast message on Finalize Block error: %s", newBlock.GetHash().String(), err.Error())
			}
		}(broadcastMsg)
	}

	return
}

func (m *Master) subMsgs() {
	go func() {
		for {
			msg, err := m.subFromP2P(m.startBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P on [Start block] error: ", err)
			}
			m.msgOnStart <- msg
		}
	}()

	go func() {
		for {
			msg, err := m.subFromP2P(m.endBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P on [End block] error: ", err)
			}
			m.msgOnEnd <- msg
		}
	}()

	go func() {
		for {
			msg, err := m.subFromP2P(m.finalizeBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P on [Finalize block] error: ", err)
			}
			m.msgOnFinalize <- msg
		}
	}()
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
