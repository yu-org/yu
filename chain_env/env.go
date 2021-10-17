package chain_env

import (
	. "github.com/yu-org/yu/state"
	. "github.com/yu-org/yu/subscribe"
	. "github.com/yu-org/yu/txpool"
	"github.com/yu-org/yu/types"
)

type ChainEnv struct {
	*StateStore
	Chain types.IBlockChain
	Base  types.IBlockBase
	Pool  ItxPool

	Sub *Subscription

	Execute func(block types.IBlock) error

	PubP2P func(topic string, msg []byte) error
	SubP2P func(topic string) ([]byte, error)
}
