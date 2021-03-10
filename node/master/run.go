package master

import (
	"net/http"
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
	newBlock := NewDefaultBlock()
	// start a new block
	err := m.land.RangeList(func(tri Tripod) error {
		return tri.StartBlock(m.chain, newBlock, m.txPool)
	})

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(newBlock)
	})
	err = m.chain.AppendBlock(newBlock)
	if err != nil {
		return err
	}

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

	err = nortifyWorkerForBlock(workersIps, StartBlockPath)
	if err != nil {
		return err
	}

	err = nortifyWorkerForBlock(workersIps, EndBlockPath)
	if err != nil {
		return err
	}

	return nortifyWorkerForBlock(workersIps, FinalizeBlockPath)
}

func nortifyWorkerForBlock(workersIps []string, path string) error {
	for _, ip := range workersIps {
		_, err := http.Get(ip + path)
		if err != nil {
			return err
		}
	}
	return nil
}
