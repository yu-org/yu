package chain_env

import (
	"github.com/yu-org/yu/p2p"
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

	Execute func(block *types.CompactBlock) error

	P2pNetwork p2p.P2pNetwork
}
