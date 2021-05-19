package env

import (
	"github.com/libp2p/go-libp2p-core/host"
	. "yu/blockchain"
	. "yu/subscribe"
	. "yu/txpool"
)

type Env struct {
	Chain IBlockChain
	Base  IBlockBase
	Pool  ItxPool

	Peer host.Host

	Sub *Subscription
}
