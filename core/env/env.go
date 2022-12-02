package env

import (
	. "github.com/yu-org/yu/core/state"
	. "github.com/yu-org/yu/core/subscribe"
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
)

type ChainEnv struct {
	State IState
	Chain IBlockChain
	TxDB  ItxDB
	Pool  ItxPool

	Sub *Subscription

	Execute func(block *Block) error

	P2pNetwork p2p.P2pNetwork
}
