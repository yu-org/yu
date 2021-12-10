package evm

import (
	gcommon "github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/types"
	"math/big"
)

// NewEVMBlockContext creates a new context for use in the EVM.
func NewEVMBlockContext(yheader *types.Header, chain *BlockChain, author *gcommon.Address) vm.BlockContext {
	header := HeaderToGeth(yheader)
	var (
		beneficiary gcommon.Address
		baseFee     *big.Int
	)

	// If we don't have an explicit author (i.e. not mining), extract from the header
	if author == nil {
		beneficiary = gcommon.Address{}
	} else {
		beneficiary = *author
	}
	if header.BaseFee != nil {
		baseFee = new(big.Int).Set(header.BaseFee)
	}
	return vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(header, chain),
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        new(big.Int).SetUint64(header.Time),
		Difficulty:  new(big.Int).Set(header.Difficulty),
		BaseFee:     baseFee,
		GasLimit:    header.GasLimit,
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *gtypes.Header, chain *BlockChain) func(n uint64) gcommon.Hash {
	// Cache will initially contain [refHash.parent],
	// Then fill up with [refHash.p, refHash.pp, refHash.ppp, ...]
	var cache []gcommon.Hash

	return func(n uint64) gcommon.Hash {
		// If there's no hash cache yet, make one
		if len(cache) == 0 {
			cache = append(cache, ref.ParentHash)
		}
		if idx := ref.Number.Uint64() - n - 1; idx < uint64(len(cache)) {
			return cache[idx]
		}
		// No luck in the cache, but we can start iterating from the last element we already know
		lastKnownHash := cache[len(cache)-1]
		lastKnownNumber := ref.Number.Uint64() - uint64(len(cache))

		for {
			block, _ := chain.GetBlock(common.Hash(lastKnownHash))
			//header := chain.GetHeader(lastKnownHash, lastKnownNumber)
			if block == nil {
				break
			}
			cache = append(cache, gcommon.Hash(block.PrevHash))
			lastKnownHash = gcommon.Hash(block.PrevHash)
			lastKnownNumber = uint64(block.Height) - 1
			if n == lastKnownNumber {
				return lastKnownHash
			}
		}
		return gcommon.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr gcommon.Address, amount *big.Int) bool {
	return db.GetBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient gcommon.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
