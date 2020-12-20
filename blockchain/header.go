package blockchain

import (
	. "yu/common"
)

type Header struct {
	parentHash Hash
	number BlockNum
	txnRoot Hash
	stateRoot Hash
}

func NewHeader(parentHash Hash, number BlockNum, txnRoot Hash, stateRoot Hash) *Header {
	return &Header{
		parentHash,
		number,
		txnRoot,
		stateRoot,
	}
}

func (h *Header) Num() BlockNum {
	return h.number
}

func (h *Header) ParentHash() Hash {
	return h.parentHash
}

func (h *Header) TxnRoot() Hash {
	return h.txnRoot
}

func (h *Header) StateRoot() Hash {
	return h.stateRoot
}
