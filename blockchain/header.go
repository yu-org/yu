package blockchain

import (
	. "yu/common"
)

type Header struct {
	parentHash string
	number BlockNum
	txnRoot string
	stateRoot string
}

func NewHeader(parentHash string, number BlockNum, txnRoot string, stateRoot string) *Header {
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

func (h *Header) ParentHash() string {
	return h.parentHash
}

func (h *Header) TRoot() string {
	return h.txnRoot
}

func (h *Header) SRoot() string {
	return h.stateRoot
}
