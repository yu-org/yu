package tripod

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/context"
	"github.com/yu-org/yu/types"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(ctx *Context, currentBlock types.IBlock) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(ctx *Context, blockHash Hash) (respObj interface{}, err error)
)
