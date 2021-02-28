package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"
	. "yu/common"
	"yu/event"
	"yu/trie"
	"yu/txn"
)

type Block struct {
	header    *Header
	txnHashes []Hash
}

func NewBlock(txnHashes []Hash, prevHash Hash, height BlockNum) IBlock {
	header := &Header{
		prevHash:  prevHash,
		number:    height + 1,
		timestamp: time.Now().UnixNano(),
	}
	return &Block{
		header:    header,
		txnHashes: txnHashes,
	}
}

func NewEmptyBlock() *Block {
	return &Block{}
}

func (b *Block) Header() IHeader {
	return b.header
}

func (b *Block) BlockId() BlockId {
	return NewBlockId(b.BlockNumber(), b.Hash())
}

func (b *Block) BlockNumber() BlockNum {
	return b.header.Num()
}

func (b *Block) Hash() Hash {
	return b.header.TxnRoot()
}

func (b *Block) SetHash(hash Hash) {
	b.header.txnRoot = hash
}

func (b *Block) StateRoot() Hash {
	return b.header.stateRoot
}

func (b *Block) SetStateRoot(hash Hash) {
	b.header.stateRoot = hash
}

func (b *Block) PrevHash() Hash {
	return b.header.PrevHash()
}

func (b *Block) Timestamp() int64 {
	return b.header.Timestamp()
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

func (b *Block) Events() []event.IEvent {
	allEvents := make([]event.IEvent, 0)
	for _, tx := range b.txns {
		events := tx.Events()
		allEvents = append(allEvents, events...)
	}
	return allEvents
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
