package chain_env

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/state"
	. "github.com/Lawliet-Chan/yu/subscribe"
	. "github.com/Lawliet-Chan/yu/txpool"
	"github.com/libp2p/go-libp2p-core/peer"
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
