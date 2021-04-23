package tripod

import (
	"yu/blockchain"
	"yu/common"
	"yu/context"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(*context.Context, blockchain.IBlock) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(*context.Context, common.Hash) (respObj interface{}, err error)
)
