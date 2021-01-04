package blockchain

import (
	"time"
	. "yu/common"
)

type Header struct {
	preHash Hash
	number     BlockNum
	txnRoot    Hash
	stateRoot  Hash
	timestamp int64
}

func NewHeader(preHash Hash, number BlockNum, txnRoot Hash, stateRoot Hash) *Header {
	timestamp := time.Now().UnixNano()
	return &Header{
		preHash,
		number,
		txnRoot,
		stateRoot,
		timestamp,
	}
}

func (h *Header) Num() BlockNum {
	return h.number
}

func (h *Header) PreHash() Hash {
	return h.preHash
}

func (h *Header) TxnRoot() Hash {
	return h.txnRoot
}

func (h *Header) StateRoot() Hash {
	return h.stateRoot
}
