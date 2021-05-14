package env

import (
	. "yu/blockchain"
	. "yu/subscribe"
	. "yu/txpool"
)

type Env struct {
	CurrentBlock IBlock
	Chain        IBlockChain
	Base         IBlockBase
	Pool         ItxPool

	Sub *Subscription
}
