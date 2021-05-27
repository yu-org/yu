package master

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/tripod"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
)

func (m *Master) Run() {

	switch m.RunMode {
	case LocalNode:
		for {
			err := m.LocalRun()
			if err != nil {
				logrus.Errorf("run blockchain error: %s", err.Error())
			}
		}
	case MasterWorker:
		for {
			err := m.MasterWokrerRun()
			logrus.Errorf("run blockchain error: %s", err.Error())
		}

	default:
		logrus.Panic(NoRunMode)
	}

}

func (m *Master) LocalRun() error {

	needBcBlock := false
	var newBlock IBlock
	// start a new block
	err := m.land.RangeList(func(tri Tripod) error {
		var (
			need bool
			err  error
		)
		newBlock, need, err = tri.StartBlock(m.GetEnv(), m.land)
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

	logrus.Infof("finish start block(%s) height(%d)", newBlock.GetHeader().GetHash().String(), newBlock.GetHeader().GetHeight())

	if needBcBlock {
		go func() {
			err := m.broadcastBlockAndTxns(newBlock)
			if err != nil {
				logrus.Errorf("broadcast block(%s) and txns error: %s", newBlock.GetHeader().GetHash().String(), err.Error())
			}
		}()
	} else {
		err = m.SyncTxns(newBlock)
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

func (m *Master) broadcastBlockAndTxns(b IBlock) error {
	blockHash := b.GetHeader().GetHash()
	txns, err := m.base.GetTxns(blockHash)
	if err != nil {
		return err
	}

	err = m.pubBlock(b)
	if err != nil {
		return err
	}
	return m.pubPackedTxns(blockHash, txns)
}
