package common

import "yu/context"

type (
	BlockNum uint64
	// Developers define the 'Execution' in the pod to let clients call.
	// Just like transactions in ETH, extrinsic in Substrate
	Execution func(ctx *context.Context) error
	// Developers define the 'Query' in the pod to let clients query the blockchain.
	// This operation has no consensus reached in the blockchain network.
	Query func(ctx *context.Context, blockNum BlockNum) error
	// The Call from clients, it is an instance of an 'Execution'.
	Call struct {
		PodName  string
		FuncName string
		Params   []interface{}
	}
)

// Lengths of hashes and addresses in bytes.
const (
	// HashLength is the expected length of the hash
	HashLen = 32
	// AddressLength is the expected length of the address
	AddressLen = 20
)

type (
	Hash      [HashLen]byte
	Address   []byte
)

func(h Hash) Bytes() []byte {
	return h[:]
}