package blockchain

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Header struct {
	PrevHash     Hash
	Hash         Hash
	Height       BlockNum
	TxnRoot      Hash
	StateRoot    Hash
	Nonce        uint64
	Timestamp    uint64
	ProducerPeer peer.ID

	LeiLimit uint64
	LeiUsed  uint64
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

func (h *Header) GetTimestamp() uint64 {
	return h.Timestamp
}

func (h *Header) GetProducerPeer() peer.ID {
	return h.ProducerPeer
}

func (h *Header) GetLeiLimit() uint64 {
	return h.LeiLimit
}

func (h *Header) GetLeiUsed() uint64 {
	return h.LeiUsed
}
