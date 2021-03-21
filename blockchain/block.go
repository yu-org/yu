package blockchain

import (
	"bytes"
	"encoding/gob"
	. "yu/common"
	"yu/trie"
	"yu/txn"
)

type Block struct {
	header     *Header
	txnsHashes []Hash
}

func (b *Block) Header() IHeader {
	return b.header
}

func (b *Block) TxnsHashes() []Hash {
	return b.txnsHashes
}

func (b *Block) SetTxnsHashes(hashes []Hash) {
	b.txnsHashes = hashes
}

func (b *Block) SetHash(hash Hash) {
	b.header.hash = hash
}

func (b *Block) SetPreHash(preHash Hash) {
	b.header.prevHash = preHash
}

func (b *Block) SetHeight(height BlockNum) {
	b.header.height = height
}

func (b *Block) BlockId() BlockId {
	return NewBlockId(b.header.Height(), b.header.Hash())
}

func (b *Block) SetStateRoot(hash Hash) {
	b.header.stateRoot = hash
}

func (b *Block) Extra() interface{} {
	return b.header.Extra()
}

func (b *Block) SetExtra(extra interface{}) {
	b.header.extra = extra
}

func (b *Block) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Block) Decode(data []byte) (IBlock, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(b)
	return b, err
}

func MakeTxnRoot(txns []txn.IsignedTxn) (Hash, error) {
	txnsBytes := make([]Hash, 0)
	for _, tx := range txns {
		hash := tx.GetTxnHash()
		txnsBytes = append(txnsBytes, hash)
	}
	mTree := trie.NewMerkleTree(txnsBytes)
	return mTree.RootNode.Data, nil
}
