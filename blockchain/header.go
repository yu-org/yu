package blockchain

import (
	. "yu/common"
)

type Header struct {
	prevHash  Hash
	hash      Hash
	number    BlockNum
	txnRoot   Hash
	stateRoot Hash
	extra     interface{}
	timestamp int64
}

func (h *Header) BlockNumber() BlockNum {
	return h.number
}

func (h *Header) Hash() Hash {
	return h.hash
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

func (h *Header) Timestamp() int64 {
	return h.timestamp
}

func (h *Header) Extra() interface{} {
	return h.extra
}
