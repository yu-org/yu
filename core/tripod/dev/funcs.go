package dev

import (
	. "github.com/yu-org/yu/core/context"
)

type (
	// Writing Developers define the 'Writing' in the pod to let clients call.
	// Just like transactions in ETH
	Writing func(ctx *WriteContext) error
	// Reading Developers define the 'Reading' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	// respObj is a json object
	Reading func(ctx *ReadContext)
	// P2pHandler is a p2p server handler. You can define the services in P2P server.
	// Just like TCP handler.
	P2pHandler func([]byte) ([]byte, error)
)
