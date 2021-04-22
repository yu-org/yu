package master

import (
	"github.com/sirupsen/logrus"
	. "yu/blockchain"
	. "yu/common"
	. "yu/node"
	. "yu/tripod"
	. "yu/yerror"
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
	newBlock := m.chain.NewDefaultBlock()
	needBcBlock := false
	// start a new block
	err := m.land.RangeList(func(tri Tripod) error {
		need, err := tri.StartBlock(m.chain, newBlock, m.txPool)
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
			err := m.broadcastBlockAndTxns(newBlock)
			if err != nil {
				logrus.Errorf("broadcast block(%s) and txns error: %s", newBlock.GetHeader().GetHash(), err.Error())
			}
		}()
	} else {
		err = m.SyncTxns(newBlock)
		if err != nil {
			return err
		}
	}

	err = m.txPool.Flush()
	if err != nil {
		return err
	}
	err = m.chain.FlushBlocksFromP2P(newBlock.GetHeader().GetHeight())
	if err != nil {
		return err
	}

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(m.chain, newBlock)
	})
	if err != nil {
		return err
	}

	go func() {
		err := ExecuteTxns(newBlock, m.base, m.land, m.sub)
		if err != nil {
			logrus.Errorf(
				"execute txns error at block(%s) : %s",
				newBlock.GetHeader().GetHash().String(),
				err.Error(),
			)
		}
	}()

	// finalize this block
	return m.land.RangeList(func(tri Tripod) error {
		return tri.FinalizeBlock(m.chain, newBlock)
	})
}

func (m *Master) MasterWokrerRun() error {
	workersIps, err := m.allWorkersIP()
	if err != nil {
		return err
	}

	newBlock := m.chain.NewDefaultBlock()

	err = m.nortifyWorker(workersIps, StartBlockPath, newBlock)
	if err != nil {
		return err
	}

	// todo: if need broadcast block,
	// m.readyBroadcastBlock(newBlock)

	err = m.SyncTxns(newBlock)
	if err != nil {
		return err
	}

	err = m.txPool.Flush()
	if err != nil {
		return err
	}

	err = m.chain.FlushBlocksFromP2P(newBlock.GetHeader().GetHeight())
	if err != nil {
		return err
	}

	err = m.nortifyWorker(workersIps, EndBlockPath, newBlock)
	if err != nil {
		return err
	}

	go func() {
		err := m.nortifyWorker(workersIps, ExecuteTxnsPath, newBlock)
		if err != nil {
			logrus.Errorf("nortify worker executing txns error: %s", err.Error())
		}
	}()

	return m.nortifyWorker(workersIps, FinalizeBlockPath, newBlock)
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
