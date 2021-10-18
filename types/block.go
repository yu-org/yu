package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/trie"
	"github.com/yu-org/yu/types/goproto"
)

type Block struct {
	*CompactBlock
	Txns SignedTxns
}

type CompactBlock struct {
	*Header
	TxnsHashes []Hash
}

func (b *CompactBlock) CopyFrom(other *CompactBlock) {
	*b = *other
}

func (b *CompactBlock) UseLei(lei uint64) {
	b.Header.LeiUsed += lei
}

func (b *CompactBlock) Encode() ([]byte, error) {
	return proto.Marshal(b.ToPb())
}

func DecodeCompactBlock(byt []byte) (*CompactBlock, error) {
	var b goproto.CompactBlock
	err := proto.Unmarshal(byt, &b)
	if err != nil {
		return nil, err
	}
	return CompactBlockFromPb(&b), nil
}

func (b *CompactBlock) ToPb() *goproto.CompactBlock {
	return &goproto.CompactBlock{
		Header:     b.Header.ToPb(),
		TxnsHashes: HashesToTwoBytes(b.TxnsHashes),
	}
}

func CompactBlockFromPb(pb *goproto.CompactBlock) *CompactBlock {
	header := HeaderFromPb(pb.Header)
	return &CompactBlock{
		Header:     header,
		TxnsHashes: TwoBytesToHashes(pb.TxnsHashes),
	}
}

func IfLeiOut(Lei uint64, block *CompactBlock) bool {
	return Lei+block.LeiUsed > block.LeiLimit
}

func MakeTxnRoot(txns []*SignedTxn) (Hash, error) {
	txnsBytes := make([]Hash, 0)
	for _, tx := range txns {
		hash := tx.GetTxnHash()
		txnsBytes = append(txnsBytes, hash)
	}
	mTree := trie.NewMerkleTree(txnsBytes)
	return mTree.RootNode.Data, nil
}

type Header struct {
	PrevHash  Hash
	Hash      Hash
	Height    BlockNum
	TxnRoot   Hash
	StateRoot Hash

	Timestamp uint64
	PeerID    peer.ID

	Extra []byte

	LeiLimit uint64
	LeiUsed  uint64

	Pubkey    []byte
	Signature []byte

	Validators     *goproto.Validators
	ProofBlockHash Hash
	ProofHeight    BlockNum
	VrfProof       []byte

	Nonce      uint64
	Difficulty uint64
}

func (h *Header) ToPb() *goproto.Header {
	return &goproto.Header{
		Hash:       h.Hash.Bytes(),
		PrevHash:   h.PrevHash.Bytes(),
		Height:     uint64(h.Height),
		TxnRoot:    h.TxnRoot.Bytes(),
		StateRoot:  h.StateRoot.Bytes(),
		Timestamp:  h.Timestamp,
		PeerID:     h.PeerID.String(),
		LeiLimit:   h.LeiLimit,
		LeiUsed:    h.LeiUsed,
		Validators: h.Validators,

		ProofBlockHash: h.ProofBlockHash.Bytes(),
		ProofHeight:    uint64(h.ProofHeight),
		VrfProof:       h.VrfProof,

		Nonce:      h.Nonce,
		Difficulty: h.Difficulty,

		Extra: h.Extra,
	}
}

func HeaderFromPb(pb *goproto.Header) *Header {
	var (
		peerID peer.ID
		err    error
	)
	if pb.PeerID == "" {
		peerID = peer.ID("")
	} else {
		peerID, err = peer.Decode(pb.PeerID)
		if err != nil {
			logrus.Panicf("peerID(%s) decode error: %v", pb.PeerID, err)
		}
	}

	return &Header{
		PrevHash:  BytesToHash(pb.PrevHash),
		Hash:      BytesToHash(pb.Hash),
		Height:    BlockNum(pb.Height),
		TxnRoot:   BytesToHash(pb.TxnRoot),
		StateRoot: BytesToHash(pb.StateRoot),
		Timestamp: pb.Timestamp,
		PeerID:    peerID,

		LeiLimit:   pb.LeiLimit,
		LeiUsed:    pb.LeiUsed,
		Validators: pb.Validators,

		ProofBlockHash: BytesToHash(pb.ProofBlockHash),
		ProofHeight:    BlockNum(pb.ProofHeight),
		VrfProof:       pb.VrfProof,

		Nonce:      pb.Nonce,
		Difficulty: pb.Difficulty,

		Extra: pb.Extra,
	}
}
