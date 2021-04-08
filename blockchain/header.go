package blockchain

import (
	. "yu/common"
)

type Header struct {
	PrevHash  Hash
	Hash      Hash
	Height    BlockNum
	TxnRoot   Hash
	StateRoot Hash
	Nonce     int64
	Timestamp int64
}

func (h *Header) GetHeight() BlockNum {
	return h.Height
}

func (h *Header) GetHash() Hash {
	return h.Hash
}

func (h *Header) GetPrevHash() Hash {
	return h.PrevHash
}

func (h *Header) GetTxnRoot() Hash {
	return h.TxnRoot
}

func (h *Header) GetStateRoot() Hash {
	return h.StateRoot
}

func (h *Header) GetTimestamp() int64 {
	return h.Timestamp
}

func (h *Header) GetNonce() int64 {
	return h.Nonce
}
