package mpt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage/kv"
	"io"
	"testing"
	"unicode/utf8"
)

func MustUnmarshal(data []byte, load interface{}) {
	err := json.Unmarshal(data, &load)
	if err != nil {
		panic(err)
	}
}

type MPTMerkleProof struct {
	RootHash  []byte   `json:"r"`
	HashChain [][]byte `json:"h"`
}

func validateMerklePatriciaTrie(
	Proof []byte,
	Key []byte,
	Value []byte,
	hfType uint8,
) bool {
	var jsonProof MPTMerkleProof
	MustUnmarshal(Proof, &jsonProof)

	var hf = Keccak256
	keybuf := bytes.NewReader(Key)

	var keyrune rune
	var keybyte byte
	// var rsize int
	var err error
	var hashChain = jsonProof.HashChain
	var curNode node
	var curHash []byte = jsonProof.RootHash
	// TODO: export node decoder
	for {

		if len(hashChain) == 0 {
			return false
		}
		if !bytes.Equal(curHash, hf(hashChain[0])) {
			return false
		}

		curNode, err = DecodeNode(curHash, hashChain[0])
		if err != nil {
			return false
		}
		hashChain = hashChain[1:]

		switch n := curNode.(type) {
		case *TrieFullNode:
			keyrune, _, err = keybuf.ReadRune()
			if err == io.EOF {
				if len(hashChain) != 0 {
					return false
				}
				cld, ok := n.Children[16].(TrieValueNode)
				if !ok {
					return false
				}
				if !bytes.Equal(cld[:], Value) {
					return false
				}
				// else:
				goto CheckKeyValueOK
			} else if err != nil {
				return false
			}
			if keyrune == utf8.RuneError {
				return false
			}
			cld, ok := n.Children[int(keyrune)].(TrieHashNode)
			if !ok {
				return false
			}
			curHash = cld[:]
		case *TrieShortNode:
			for idx := 0; idx < len(n.Key); idx++ {
				keybyte, err = keybuf.ReadByte()
				if err == io.EOF {
					if idx != len(n.Key)-1 {
						if Value != nil {
							return false
						} else {
							goto CheckKeyValueOK
						}
					} else {
						if len(hashChain) != 0 {
							return false
						}
						cld, ok := n.Val.(TrieValueNode)
						if !ok {
							return false
						}
						if !bytes.Equal(cld[:], Value) {
							return false
						}
						// else:
						goto CheckKeyValueOK
					}
				} else if err != nil {
					return false
				}
				if keybyte != n.Key[idx] {
					return Value == nil
				}
			}
			cld, ok := n.Val.(TrieValueNode)
			if !ok {
				return false
			}
			curHash = cld[:]
		}
	}
CheckKeyValueOK:

	return true
}

func TestGenerateProof(t *testing.T) {
	cfg := &config.KVconf{
		KvType: "badger",
		Path:   "./testdb",
	}
	kvdb, err := kv.NewKvdb(cfg)
	if err != nil {
		assert.NoError(t, err)
	}
	db := NewNodeBase(kvdb)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	var tr *Trie
	tr, err = NewTrie(HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), db)
	if err != nil {
		t.Error(err)
		return
	}

	tr.Update([]byte("key"), []byte("..."))
	tr.Update([]byte("keyy"), []byte("..."))
	tr.Update([]byte("keyyyy"), []byte("..."))

	var trHash Hash
	trHash, err = tr.Commit(nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(trHash)
	tr, err = NewTrie(trHash, db)
	var proof [][]byte
	fmt.Println("--------------------------------")
	proof, err = tr.TryProve([]byte("keyy"))
	for _, bt := range proof {
		fmt.Println(hex.EncodeToString(bt))
	}
}

func TestGenerateLongProof(t *testing.T) {
	cfg := &config.KVconf{
		KvType: "badger",
		Path:   "./testdb",
	}
	kvdb, err := kv.NewKvdb(cfg)
	if err != nil {
		assert.NoError(t, err)
	}
	db := NewNodeBase(kvdb)
	defer db.Close()
	var tr *Trie
	tr, err = NewTrie(HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"), db)
	if err != nil {
		t.Error(err)
		return
	}

	tr.Update([]byte("\x20\x20\x20\x20\x20\x20\x20\x20"), []byte("..."))
	tr.Update([]byte("\x20\x20\x20\x20"), []byte("..."))
	tr.Update([]byte("\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x21"), []byte("..."))

	var trHash Hash
	trHash, err = tr.Commit(nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(trHash)
	tr, err = NewTrie(trHash, db)
	var proof [][]byte
	fmt.Println("--------------------------------")
	proof, err = tr.TryProve([]byte("keyy"))
	for _, bt := range proof {
		fmt.Println(hex.EncodeToString(bt))
	}
	fmt.Println("--------------------------------")
	proof, err = tr.TryProve([]byte("\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x21"))
	for _, bt := range proof {
		fmt.Println(hex.EncodeToString(bt))
	}
}
