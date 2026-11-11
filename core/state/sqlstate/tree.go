// Package sqlstate stores table state and pending forks in one SQL database.
// Each database must have a single execution owner. Models and their encoding
// are protocol configuration and must agree across nodes.
package sqlstate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type digest [sha256.Size]byte

type storedNode struct {
	Hash string `gorm:"primaryKey;size:64"`
	Data []byte
}

// TableName identifies the internal immutable node table.
func (storedNode) TableName() string { return "_yu_state_nodes" }

type storedRow struct {
	Hash string `gorm:"primaryKey;size:64"`
	Data []byte
}

// TableName identifies the internal immutable row table.
func (storedRow) TableName() string { return "_yu_state_rows" }

type rowData struct {
	Table string
	Key   []byte
	Data  []byte
}

var emptyRoots = makeEmptyRoots()

// makeEmptyRoots computes the default hash at every depth of the 256-bit tree.
func makeEmptyRoots() [257]digest {
	var roots [257]digest
	empty := sha256.Sum256([]byte{0})
	for depth := range roots {
		roots[depth] = empty
	}
	return roots
}

// branchBytes encodes a branch with fixed-width, ordered child hashes.
func branchBytes(left, right digest) []byte {
	data := make([]byte, 1, 65)
	data[0] = 2
	data = append(data, left[:]...)
	return append(data, right[:]...)
}

// pathFor hashes the unambiguous table/key tuple into a tree path.
func pathFor(table string, key []byte) digest {
	data, _ := json.Marshal(struct {
		Table string
		Key   []byte
	}{table, key})
	return sha256.Sum256(append([]byte{4}, data...))
}

// parseDigest validates a persisted hexadecimal hash.
func parseDigest(value string) (digest, error) {
	var result digest
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("invalid state hash %q", value)
	}
	copy(result[:], decoded)
	return result, nil
}

// hexDigest returns the portable SQL primary key for a hash.
func hexDigest(value digest) string { return hex.EncodeToString(value[:]) }

// loadNode reads and verifies an immutable node; empty nodes are never stored.
func loadNode(db *gorm.DB, hash digest) ([]byte, error) {
	var node storedNode
	if err := db.First(&node, "hash = ?", hexDigest(hash)).Error; err != nil {
		return nil, fmt.Errorf("reading state node: %w", err)
	}
	if sha256.Sum256(node.Data) != hash {
		return nil, errors.New("corrupt state node")
	}
	return node.Data, nil
}

// putNode stores an immutable node, reusing identical content.
func putNode(db *gorm.DB, data []byte) (digest, error) {
	hash := sha256.Sum256(data)
	err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&storedNode{Hash: hexDigest(hash), Data: data}).Error
	return hash, err
}

// children decodes a branch. The caller must only pass non-empty branch nodes.
func children(data []byte) (digest, digest, error) {
	var left, right digest
	if len(data) != 65 || data[0] != 2 {
		return left, right, errors.New("invalid state branch")
	}
	copy(left[:], data[1:33])
	copy(right[:], data[33:65])
	return left, right, nil
}

// loadRow verifies a content-addressed row before decoding its envelope.
func loadRow(db *gorm.DB, hash digest) (rowData, error) {
	var result rowData
	var row storedRow
	if err := db.First(&row, "hash = ?", hexDigest(hash)).Error; err != nil {
		return result, fmt.Errorf("reading state row: %w", err)
	}
	if sha256.Sum256(append([]byte{3}, row.Data...)) != hash {
		return result, errors.New("corrupt state row")
	}
	err := json.Unmarshal(row.Data, &result)
	return result, err
}

// leafRow verifies a leaf and its row identity at depth 256.
func leafRow(db *gorm.DB, data []byte) (rowData, error) {
	if len(data) != 65 || data[0] != 1 {
		return rowData{}, errors.New("invalid state leaf")
	}
	var path, rowHash digest
	copy(path[:], data[1:33])
	copy(rowHash[:], data[33:65])
	row, err := loadRow(db, rowHash)
	if err == nil && pathFor(row.Table, row.Key) != path {
		err = errors.New("state leaf identity mismatch")
	}
	return row, err
}

// treeGet reads a row from a specific root, never from a latest-value cache.
func treeGet(db *gorm.DB, root digest, table string, key []byte) ([]byte, error) {
	path := pathFor(table, key)
	for depth := 0; depth <= 256; depth++ {
		if root == emptyRoots[depth] {
			return nil, ErrNotFound
		}
		data, err := loadNode(db, root)
		if err != nil {
			return nil, err
		}
		if len(data) == 65 && data[0] == 1 {
			row, err := leafRow(db, data)
			if err != nil {
				return nil, err
			}
			if row.Table != table || !bytes.Equal(row.Key, key) {
				return nil, ErrNotFound
			}
			return row.Data, nil
		}
		if depth == 256 {
			return nil, errors.New("state branch exceeds path depth")
		}
		left, right, err := children(data)
		if err != nil {
			return nil, err
		}
		root = left
		if path[depth/8]&(1<<uint(7-depth%8)) != 0 {
			root = right
		}
	}
	panic("unreachable tree depth")
}

// treeUpdate copies only the affected path. Existing nodes remain immutable.
func treeUpdate(db *gorm.DB, root digest, row rowData, deleted bool) (digest, error) {
	path := pathFor(row.Table, row.Key)
	leaf := emptyRoots[256]
	if !deleted {
		encoded, err := json.Marshal(row)
		if err != nil {
			return digest{}, err
		}
		rowHash := sha256.Sum256(append([]byte{3}, encoded...))
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&storedRow{Hash: hexDigest(rowHash), Data: encoded}).Error; err != nil {
			return digest{}, err
		}
		data := append([]byte{1}, path[:]...)
		leaf, err = putNode(db, append(data, rowHash[:]...))
		if err != nil {
			return digest{}, err
		}
	}
	return updatePath(db, root, 0, path, leaf, deleted)
}

// updatePath uses compressed terminal paths: a lone leaf occupies its subtree
// directly. Branches retain bit positions, and deleting a sibling bubbles up a
// lone leaf without changing any other pending version's immutable nodes.
func updatePath(db *gorm.DB, root digest, depth int, path, leaf digest, deleted bool) (digest, error) {
	if root == emptyRoots[depth] {
		return leaf, nil
	}
	data, err := loadNode(db, root)
	if err != nil {
		return digest{}, err
	}
	if len(data) == 65 && data[0] == 1 {
		var prior digest
		copy(prior[:], data[1:33])
		if prior == path {
			return leaf, nil
		}
		if deleted {
			return root, nil
		}
		split := depth
		for split < 256 && pathBit(prior, split) == pathBit(path, split) {
			split++
		}
		if split == 256 {
			return digest{}, errors.New("state key collision")
		}
		left, right := root, leaf
		if pathBit(path, split) == 0 {
			left, right = leaf, root
		}
		next, err := putNode(db, branchBytes(left, right))
		if err != nil {
			return digest{}, err
		}
		for bit := split - 1; bit >= depth; bit-- {
			left, right = next, emptyRoots[bit+1]
			if pathBit(path, bit) != 0 {
				left, right = right, left
			}
			next, err = putNode(db, branchBytes(left, right))
			if err != nil {
				return digest{}, err
			}
		}
		return next, nil
	}
	if depth == 256 {
		return digest{}, errors.New("state branch exceeds path depth")
	}
	left, right, err := children(data)
	if err != nil {
		return digest{}, err
	}
	if pathBit(path, depth) == 0 {
		left, err = updatePath(db, left, depth+1, path, leaf, deleted)
	} else {
		right, err = updatePath(db, right, depth+1, path, leaf, deleted)
	}
	if err != nil {
		return digest{}, err
	}
	if left == emptyRoots[depth+1] && right == emptyRoots[depth+1] {
		return emptyRoots[depth], nil
	}
	if left == emptyRoots[depth+1] || right == emptyRoots[depth+1] {
		remaining := left
		if remaining == emptyRoots[depth+1] {
			remaining = right
		}
		child, err := loadNode(db, remaining)
		if err != nil {
			return digest{}, err
		}
		if len(child) == 65 && child[0] == 1 {
			return remaining, nil
		}
	}
	return putNode(db, branchBytes(left, right))
}

// pathBit returns a most-significant-first bit; bit must be in [0, 256).
func pathBit(path digest, bit int) byte { return (path[bit/8] >> uint(7-bit%8)) & 1 }

// walkTree visits reachable rows and optionally marks nodes and blobs for GC.
func walkTree(db *gorm.DB, root digest, depth int, nodes, rows map[string]bool, visit func(rowData) error) error {
	if root == emptyRoots[depth] || nodes[hexDigest(root)] {
		return nil
	}
	data, err := loadNode(db, root)
	if err != nil {
		return err
	}
	nodes[hexDigest(root)] = true
	if len(data) == 65 && data[0] == 1 {
		row, err := leafRow(db, data)
		if err != nil {
			return err
		}
		rows[hex.EncodeToString(data[33:65])] = true
		if visit != nil {
			return visit(row)
		}
		return nil
	}
	if depth == 256 {
		return errors.New("state branch exceeds path depth")
	}
	left, right, err := children(data)
	if err != nil {
		return err
	}
	if err := walkTree(db, left, depth+1, nodes, rows, visit); err != nil {
		return err
	}
	return walkTree(db, right, depth+1, nodes, rows, visit)
}
