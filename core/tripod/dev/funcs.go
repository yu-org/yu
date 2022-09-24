package dev

import (
	. "github.com/yu-org/yu/core/context"
)

type (
	// Execution Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(ctx *Context) error
	// Query Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Query func(ctx *Context) (respObj interface{}, err error)
	// P2pHandler is a p2p server handler. You can define the services in P2P server.
	// Just like TCP handler.
	P2pHandler func([]byte) ([]byte, error)
)
