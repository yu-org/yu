package chain_env

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/subscribe"
	. "github.com/Lawliet-Chan/yu/txpool"
	"github.com/libp2p/go-libp2p-core/host"
)

type ChainEnv struct {
	Chain IBlockChain
	Base  IBlockBase
	Pool  ItxPool

	Peer host.Host

	Sub *Subscription
}
