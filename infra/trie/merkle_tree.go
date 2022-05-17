package trie

import (
	"crypto/sha256"
	. "github.com/yu-org/yu/common"
)

// MerkleTree represent a Merkle tree
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode represent a Merkle tree node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  Hash
}

// NewMerkleTree creates a new Merkle tree from a sequence of data
func NewMerkleTree(hashes []Hash) *MerkleTree {
	if len(hashes) == 0 {
		return &MerkleTree{RootNode: &MerkleNode{
			Left:  nil,
			Right: nil,
			Data:  NullHash,
		}}
	}

	var nodes []*MerkleNode

	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}

	for _, hash := range hashes {
		leaf := newMerkleNode(nil, nil, hash)
		nodes = append(nodes, leaf)
	}

	for {
		var newLevel []*MerkleNode

		for j := 0; j < len(nodes)-1; j += 2 {

			node := newMerkleNode(nodes[j], nodes[j+1], NullHash)
			newLevel = append(newLevel, node)
		}
		nodes = newLevel

		if len(nodes) == 1 {
			return &MerkleTree{RootNode: nodes[0]}
		}
	}
}

// NewMerkleNode creates a new Merkle tree node
func newMerkleNode(left, right *MerkleNode, defaultHash Hash) *MerkleNode {
	mNode := &MerkleNode{}

	if left == nil && right == nil {
		mNode.Data = sha256.Sum256(defaultHash.Bytes())
	} else {
		prevHashes := append(left.Data.Bytes(), right.Data.Bytes()...)
		mNode.Data = sha256.Sum256(prevHashes)
	}

	mNode.Left = left
	mNode.Right = right

	return mNode
}
