package chain_env

import (
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/state"
	. "github.com/yu-org/yu/subscribe"
	. "github.com/yu-org/yu/txpool"
)

type ChainEnv struct {
	*StateStore
	Chain IBlockChain
	Base  IBlockBase
	Pool  ItxPool

	Sub *Subscription

	Execute func(block IBlock) error

	PubP2P func(topic string, msg []byte) error
	SubP2P func(topic string) ([]byte, error)
}
