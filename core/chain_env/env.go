package chain_env

import (
	. "github.com/yu-org/yu/core/state"
	. "github.com/yu-org/yu/core/subscribe"
	. "github.com/yu-org/yu/core/txpool"
	types2 "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
)

type ChainEnv struct {
	*StateStore
	Chain types2.IBlockChain
	Base  types2.IBlockBase
	Pool  ItxPool

	Sub *Subscription

	Execute func(block *types2.CompactBlock) error

	P2pNetwork p2p.P2pNetwork
}
