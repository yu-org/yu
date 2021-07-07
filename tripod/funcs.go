package tripod

import (
	"github.com/Lawliet-Chan/yu/chain_env"
	"github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/context"
)

type (
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(*context.Context, *chain_env.ChainEnv) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(*context.Context, *chain_env.ChainEnv, common.Hash) (respObj interface{}, err error)
)
