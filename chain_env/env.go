package chain_env

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/state"
	. "github.com/yu-org/yu/subscribe"
	. "github.com/yu-org/yu/txpool"
)

type ChainEnv struct {
	*StateStore
	RunMode RunMode
	Chain   IBlockChain
	Base    IBlockBase
	Pool    ItxPool

	P2pID peer.ID

	Sub *Subscription
}
