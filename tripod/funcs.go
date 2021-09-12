package tripod

import (
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/chain_env"
	. "github.com/yu-altar/yu/common"
	. "github.com/yu-altar/yu/context"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(ctx *Context, currentBlock IBlock, env *ChainEnv) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(ctx *Context, env *ChainEnv, blockHash Hash) (respObj interface{}, err error)
)
