package blockchain

import (
	"time"
	. "yu/common"
)

type Header struct {
	prevHash  Hash
	number    BlockNum
	txnRoot   Hash
	stateRoot Hash
	extra     interface{}
	timestamp int64
}

func NewHeader(prevHash Hash, number BlockNum, txnRoot Hash, stateRoot Hash, extra interface{}) *Header {
	timestamp := time.Now().UnixNano()
	return &Header{
		prevHash,
		number,
		txnRoot,
		stateRoot,
		extra,
		timestamp,
	}
}

func (h *Header) Num() BlockNum {
	return h.number
}

func (h *Header) PrevHash() Hash {
	return h.prevHash
}

func (h *Header) TxnRoot() Hash {
	return h.txnRoot
}

func (h *Header) StateRoot() Hash {
	return h.stateRoot
}

func (h *Header) Extra() interface{} {
	return h.extra
}
