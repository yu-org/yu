package master

import (
	"net/http"
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

	// todo: execute txns

	// end block and append to chain
	err = m.land.RangeList(func(tri Tripod) error {
		return tri.EndBlock(m.chain, newBlock)
	})
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

	err = nortifyWorker(workersIps, StartBlockPath)
	if err != nil {
		return err
	}

	err = nortifyWorker(workersIps, ExecuteTxnsPath)
	if err != nil {
		return err
	}

	err = nortifyWorker(workersIps, EndBlockPath)
	if err != nil {
		return err
	}

	return nortifyWorker(workersIps, FinalizeBlockPath)
}

func nortifyWorker(workersIps []string, path string) error {
	for _, ip := range workersIps {
		_, err := http.Get(ip + path)
		if err != nil {
			return err
		}
	}
	return nil
}
