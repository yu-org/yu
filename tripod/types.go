package tripod

import (
	"yu/common"
	"yu/context"
	"yu/env"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(*context.Context, *env.Env) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(*context.Context, common.Hash, *env.Env) (respObj interface{}, err error)
)
