package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"sort"

	"github.com/yu-org/yu/common"
)

// ComputeHeaderHash hashes a sealed header without its hash or signature.
// Transaction contents are committed through TxnRoot, not encoded a second time.
func ComputeHeaderHash(block *Block) (common.Hash, error) {
	if block == nil || block.Header == nil {
		return common.NullHash, errors.New("missing block header")
	}
	header := *block.Header
	header.Hash, header.MinerSignature = common.NullHash, nil
	data, err := (&Block{Header: &header}).Encode()
	if err != nil {
		return common.NullHash, err
	}
	return common.BytesToHash(common.Sha256(data)), nil
}

// CalculateExecutionReceiptRoot commits to deterministic receipt fields in hash
// order. Local save times and the final block hash cannot participate in sealing.
func CalculateExecutionReceiptRoot(receipts map[common.Hash]*Receipt) (common.Hash, error) {
	keys := make([]common.Hash, 0, len(receipts))
	for key := range receipts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return bytes.Compare(keys[i][:], keys[j][:]) < 0 })
	leaves := make([]common.Hash, 0, len(keys))
	for _, key := range keys {
		if receipts[key] == nil {
			return common.NullHash, errors.New("nil execution receipt")
		}
		receipt := *receipts[key]
		receipt.BlockHash, receipt.SaveTimestamp = common.NullHash, 0
		encoded, err := receipt.Encode()
		if err != nil {
			return common.NullHash, err
		}
		hash := sha256.Sum256(append([]byte{16}, encoded...))
		leaves = append(leaves, common.BytesToHash(hash[:]))
	}
	if len(leaves) == 0 {
		return common.NullHash, nil
	}
	for len(leaves) > 1 {
		next := make([]common.Hash, 0, (len(leaves)+1)/2)
		for i := 0; i < len(leaves); i += 2 {
			right := leaves[i]
			if i+1 < len(leaves) {
				right = leaves[i+1]
			}
			data := append([]byte{17}, leaves[i][:]...)
			data = append(data, right[:]...)
			hash := sha256.Sum256(data)
			next = append(next, common.BytesToHash(hash[:]))
		}
		leaves = next
	}
	return leaves[0], nil
}

// VerifyHeaderCommitments checks the claimed header and transaction commitments.
// Signature verification and execution-root verification are separate steps.
func VerifyHeaderCommitments(block *Block) error {
	hash, err := ComputeHeaderHash(block)
	if err != nil {
		return err
	}
	if hash != block.Hash {
		return errors.New("block header hash mismatch")
	}
	root, err := MakeTxnRoot(block.Txns)
	if err != nil {
		return err
	}
	if root != block.TxnRoot {
		return errors.New("block transaction root mismatch")
	}
	return nil
}
