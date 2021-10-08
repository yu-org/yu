package blockchain

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/common"
)

type Header struct {
	PrevHash  Hash
	Hash      Hash
	Height    BlockNum
	TxnRoot   Hash
	StateRoot Hash
	Nonce     uint64
	Timestamp uint64
	PeerID    peer.ID

	Pubkey    []byte
	Signature []byte

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

func (h *Header) GetPeerID() peer.ID {
	return h.PeerID
}

func (h *Header) GetLeiLimit() uint64 {
	return h.LeiLimit
}

func (h *Header) GetLeiUsed() uint64 {
	return h.LeiUsed
}

func (h *Header) GetSignature() []byte {
	return h.Signature
}

//
//func (h *Header) GetSign() []byte {
//	return h.Signature
//}
//
//func (h *Header) GetPubkey() PubKey {
//	if h.Pubkey == nil {
//		return nil
//	}
//	pubkey, err := PubKeyFromBytes(h.Pubkey)
//	if err != nil {
//		logrus.Panic("get pubkey from block-header error: ", err.Error())
//	}
//	return pubkey
//}
