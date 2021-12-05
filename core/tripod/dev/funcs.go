package dev

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/types"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(ctx *Context, currentBlock *types.CompactBlock) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(ctx *Context, blockHash Hash) (respObj interface{}, err error)

	P2pHandler func([]byte) ([]byte, error)
)
