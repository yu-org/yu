package master

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	"github.com/Lawliet-Chan/yu/config"
	. "github.com/Lawliet-Chan/yu/state"
	. "github.com/Lawliet-Chan/yu/txpool"
	"github.com/sirupsen/logrus"
)

func loadComponents(cfg *config.MasterConf) (IBlockChain, IBlockBase, *StateStore, ItxPool) {
	chain, err := NewBlockChain(&cfg.BlockChain)
	if err != nil {
		logrus.Panicf("load blockchain error: %s", err.Error())
	}
	base, err := NewBlockBase(&cfg.BlockBase)
	if err != nil {
		logrus.Panicf("load blockbase error: %s", err.Error())
	}
	stateStore, err := NewStateStore(&cfg.State)
	if err != nil {
		logrus.Panicf("load stateKV error: %s", err.Error())
	}
	pool := LocalWithDefaultChecks(&cfg.Txpool)
	return chain, base, stateStore, pool
}
