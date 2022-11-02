// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package mpt

import (
	"fmt"
	"github.com/yu-org/yu/common"
	"io"
	"strings"

	"github.com/HyperService-Consortium/go-rlp"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	fstring(string) string
	cache() (TrieHashNode, bool)
}

type (
	TrieFullNode struct {
		Children [17]node // Actual trie node data to encode/decode (needs custom encoder)
		flags    nodeFlag
	}
	TrieShortNode struct {
		Key   []byte
		Val   node
		flags nodeFlag
	}
	TrieHashNode  []byte
	TrieValueNode []byte
)

// nilValueNode is used when collapsing internal trie nodes for hashing, since
// unset children need to serialize correctly.
var nilValueNode = TrieValueNode(nil)

// EncodeRLP encodes a full node into the consensus RLP format.
func (n *TrieFullNode) EncodeRLP(w io.Writer) error {
	var nodes [17]node

	for i, child := range &n.Children {
		if child != nil {
			nodes[i] = child
		} else {
			nodes[i] = nilValueNode
		}
	}
	return rlp.Encode(w, nodes)
}

func (n *TrieFullNode) copy() *TrieFullNode   { copy := *n; return &copy }
func (n *TrieShortNode) copy() *TrieShortNode { copy := *n; return &copy }

// nodeFlag contains caching-related metadata about a node.
type nodeFlag struct {
	hash  TrieHashNode // cached hash of the node (may be nil)
	dirty bool         // whether the node has changes that must be written to the database
}

func (n *TrieFullNode) cache() (TrieHashNode, bool)  { return n.flags.hash, n.flags.dirty }
func (n *TrieShortNode) cache() (TrieHashNode, bool) { return n.flags.hash, n.flags.dirty }
func (n TrieHashNode) cache() (TrieHashNode, bool)   { return nil, true }
func (n TrieValueNode) cache() (TrieHashNode, bool)  { return nil, true }

// Pretty printing.
func (n *TrieFullNode) String() string  { return n.fstring("") }
func (n *TrieShortNode) String() string { return n.fstring("") }
func (n TrieHashNode) String() string   { return n.fstring("") }
func (n TrieValueNode) String() string  { return n.fstring("") }

func (n *TrieFullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ", ind)
	for i, node := range &n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil> ", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind+"  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}
func (n *TrieShortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind+"  "))
}
func (n TrieHashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}
func (n TrieValueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

func mustDecodeNode(hash, buf []byte) node {
	n, err := DecodeNode(hash, buf)
	if err != nil {
		panic(fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

// DecodeNode parses the RLP encoding of a trie node.
func DecodeNode(hash, buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	elems, _, err := rlp.SplitList(buf)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	switch c, _ := rlp.CountValues(elems); c {
	case 2:
		n, err := decodeShort(hash, elems)
		return n, wrapError(err, "short")
	case 17:
		n, err := decodeFull(hash, elems)
		return n, wrapError(err, "full")
	default:
		return nil, fmt.Errorf("invalid number of list elements: %v", c)
	}
}

func DecodeNodeLazy(hash, buf []byte) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	elems, _, err := rlp.SplitList(buf)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	switch c, _ := rlp.CountValues(elems); c {
	case 2:
		n, err := decodeShort(hash, elems)
		return n, wrapError(err, "short")
	case 17:
		n, err := decodeFull(hash, elems)
		return n, wrapError(err, "full")
	default:
		return nil, fmt.Errorf("invalid number of list elements: %v", c)
	}
}

func decodeShort(hash, elems []byte) (node, error) {
	// fmt.Println("...........")
	// fmt.Println(hex.EncodeToString(hash), hex.EncodeToString(elems))
	kbuf, rest, err := rlp.SplitString(elems)
	if err != nil {
		return nil, err
	}
	flag := nodeFlag{hash: hash}
	key := compactToHex(kbuf)
	if hasTerm(key) {
		// value node
		// fmt.Println("value", key, hex.EncodeToString(kbuf), hex.EncodeToString(rest))
		val, _, err := rlp.SplitString(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid value node: %v", err)
		}
		return &TrieShortNode{key, append(TrieValueNode{}, val...), flag}, nil
	}
	// fmt.Println("novalue", key, hex.EncodeToString(kbuf), hex.EncodeToString(rest))
	r, _, err := decodeRef(rest)
	if err != nil {
		return nil, wrapError(err, "val")
	}
	return &TrieShortNode{key, r, flag}, nil
}

func decodeFull(hash, elems []byte) (*TrieFullNode, error) {
	n := &TrieFullNode{flags: nodeFlag{hash: hash}}
	for i := 0; i < 16; i++ {
		cld, rest, err := decodeRef(elems)
		if err != nil {
			return n, wrapError(err, fmt.Sprintf("[%d]", i))
		}
		n.Children[i], elems = cld, rest
	}
	val, _, err := rlp.SplitString(elems)
	if err != nil {
		return n, err
	}
	if len(val) > 0 {
		n.Children[16] = append(TrieValueNode{}, val...)
	}
	return n, nil
}

const hashLen = len(common.Hash{})

func decodeRef(buf []byte) (node, []byte, error) {
	kind, val, rest, err := rlp.Split(buf)
	if err != nil {
		return nil, buf, err
	}
	switch {
	case kind == rlp.List:
		// 'embedded' node reference. The encoding must be smaller
		// than a hash in order to be valid.
		if size := len(buf) - len(rest); size > hashLen {
			err := fmt.Errorf("oversized embedded node (size is %d bytes, want size < %d)", size, hashLen)
			return nil, buf, err
		}
		n, err := DecodeNode(nil, buf)
		return n, rest, err
	case kind == rlp.String && len(val) == 0:
		// empty node
		return nil, rest, nil
	case kind == rlp.String && len(val) == 32:
		return append(TrieHashNode{}, val...), rest, nil
	default:
		return nil, nil, fmt.Errorf("invalid RLP string size %d (want 0 or 32)", len(val))
	}
}

// wraps a decoding error with information about the path to the
// invalid child node (for debugging encoding issues).
type decodeError struct {
	what  error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err == nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError{err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}
