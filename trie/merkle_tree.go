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
func NewMerkleTree(data []Hash) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{RootNode: &MerkleNode{
			Left:  nil,
			Right: nil,
			Data:  NullHash,
		}}
	}

	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := newMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := newMerkleNode(&nodes[j], &nodes[j+1], NullHash)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

// NewMerkleNode creates a new Merkle tree node
func newMerkleNode(left, right *MerkleNode, data Hash) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		mNode.Data = sha256.Sum256(data.Bytes())
	} else {
		prevHashes := append(left.Data.Bytes(), right.Data.Bytes()...)
		mNode.Data = sha256.Sum256(prevHashes)
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}
