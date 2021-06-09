package blockchain

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/trie"
	"github.com/Lawliet-Chan/yu/txn"
	. "github.com/Lawliet-Chan/yu/utils/codec"
)

type Block struct {
	Header     *Header
	TxnsHashes []Hash
}

func (b *Block) GetHeight() BlockNum {
	return b.Header.GetHeight()
}

func (b *Block) GetHash() Hash {
	return b.Header.GetHash()
}

func (b *Block) GetPrevHash() Hash {
	return b.Header.GetPrevHash()
}

func (b *Block) GetTxnRoot() Hash {
	return b.Header.GetTxnRoot()
}

func (b *Block) GetStateRoot() Hash {
	return b.Header.GetStateRoot()
}

func (b *Block) GetTimestamp() uint64 {
	return b.Header.GetTimestamp()
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

func (b *Block) SetTxnHash(hash Hash) {
	b.Header.TxnRoot = hash
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

func (b *Block) SetNonce(nonce uint64) {
	b.Header.Nonce = nonce
}

func (b *Block) Encode() ([]byte, error) {
	return GlobalCodec.EncodeToBytes(b)
}

func (b *Block) Decode(data []byte) (IBlock, error) {
	var block Block
	err := GlobalCodec.DecodeBytes(data, &block)
	return &block, err
}

func (b *Block) CopyFrom(other IBlock) {
	otherBlock := other.(*Block)
	*b = *otherBlock
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
