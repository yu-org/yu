package master

import (
	"github.com/sirupsen/logrus"
	. "yu/blockchain"
	. "yu/common"
	. "yu/node"
	. "yu/tripod"
	. "yu/yerror"
)

func (m *Master) Run() error {
	switch m.RunMode {
	case LocalNode:
		return m.LocalRun()
	case MasterWorker:
		return m.MasterWokrerRun()
	default:
		return NoRunMode
	}
}

func (m *Master) LocalRun() error {
	newBlock := m.chain.NewDefaultBlock()
	// start a new block
	err := m.land.RangeList(func(tri Tripod) error {
		return tri.StartBlock(m.chain, newBlock, m.txPool)
	})
	if err != nil {
		return err
	}

	go m.readyBroadcastBlock(newBlock)

	err = m.SyncTxns(newBlock)
	if err != nil {
		return err
	}

	err = m.txPool.Flush()
	if err != nil {
		return err
	}
	err = m.chain.FlushBlocksFromP2P(newBlock.Header().Height())
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
		err := ExecuteTxns(newBlock, m.base, m.land)
		if err != nil {
			logrus.Errorf(
				"execute txns error at block(%s) : %s",
				newBlock.Header().Hash().String(),
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

	go m.readyBroadcastBlock(newBlock)

	err = m.SyncTxns(newBlock)
	if err != nil {
		return err
	}

	err = m.txPool.Flush()
	if err != nil {
		return err
	}

	err = m.chain.FlushBlocksFromP2P(newBlock.Header().Height())
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

func (m *Master) readyBroadcastBlock(b IBlock) {
	tbody, err := NewBlockTransferBody(b)
	if err != nil {
		logrus.Errorf("ready broadcast block(%s) error: %s", b.Header().Hash().String(), err.Error())
	}
	m.blockBcChan <- tbody
}
