package chain_env

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/common"
	. "github.com/yu-altar/yu/state"
	. "github.com/yu-altar/yu/subscribe"
	. "github.com/yu-altar/yu/txpool"
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
