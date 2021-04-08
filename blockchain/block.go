package blockchain

import (
	"bytes"
	"encoding/gob"
	. "yu/common"
	"yu/trie"
	"yu/txn"
)

type Block struct {
	Header     *Header
	TxnsHashes []Hash
}

func (b *Block) GetHeader() IHeader {
	return b.Header
}

func (b *Block) GetTxnsHashes() []Hash {
	return b.TxnsHashes
}

func (b *Block) SetTxnsHashes(hashes []Hash) {
	b.TxnsHashes = hashes
}

func (b *Block) SetHash(hash Hash) {
	b.Header.Hash = hash
}

func (b *Block) SetPreHash(preHash Hash) {
	b.Header.PrevHash = preHash
}

func (b *Block) SetHeight(height BlockNum) {
	b.Header.Height = height
}

func (b *Block) GetBlockId() BlockId {
	return NewBlockId(b.Header.GetHeight(), b.Header.GetHash())
}

func (b *Block) SetStateRoot(hash Hash) {
	b.Header.StateRoot = hash
}

func (b *Block) SetNonce(nonce int64) {
	b.Header.Nonce = nonce
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

func MakeTxnRoot(txns []*txn.SignedTxn) (Hash, error) {
	txnsBytes := make([]Hash, 0)
	for _, tx := range txns {
		hash := tx.GetTxnHash()
		txnsBytes = append(txnsBytes, hash)
	}
	mTree := trie.NewMerkleTree(txnsBytes)
	return mTree.RootNode.Data, nil
}
