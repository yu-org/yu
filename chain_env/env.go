package chain_env

import (
	"github.com/libp2p/go-libp2p-core/host"
	. "yu/blockchain"
	. "yu/subscribe"
	. "yu/txpool"
)

type ChainEnv struct {
	Chain IBlockChain
	Base  IBlockBase
	Pool  ItxPool

	Peer host.Host

	Sub *Subscription
}
