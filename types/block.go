package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
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
	return CompactBlockFromPb(&b)
}

func (b *CompactBlock) ToPb() *goproto.CompactBlock {
	return &goproto.CompactBlock{
		Header:     b.Header.ToPb(),
		TxnsHashes: HashesToTwoBytes(b.TxnsHashes),
	}
}

func CompactBlockFromPb(pb *goproto.CompactBlock) (*CompactBlock, error) {
	header, err := HeaderFromPb(pb.Header)
	if err != nil {
		return nil, err
	}
	return &CompactBlock{
		Header:     header,
		TxnsHashes: TwoBytesToHashes(pb.TxnsHashes),
	}, nil
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

	Validators *goproto.Validators
	Proof      *Proof
	PowInfo    *goproto.PowInfo
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
		Proof:      h.Proof.ToPb(),
		PowInfo:    h.PowInfo,
		Extra:      h.Extra,
	}
}

func HeaderFromPb(pb *goproto.Header) (*Header, error) {
	peerID := peer.ID(pb.PeerID)
	err := peerID.Validate()
	if err != nil {
		return nil, err
	}

	return &Header{
		PrevHash:   BytesToHash(pb.PrevHash),
		Hash:       BytesToHash(pb.Hash),
		Height:     BlockNum(pb.Height),
		TxnRoot:    BytesToHash(pb.TxnRoot),
		StateRoot:  BytesToHash(pb.StateRoot),
		Timestamp:  pb.Timestamp,
		PeerID:     peerID,
		Extra:      pb.Extra,
		LeiLimit:   pb.LeiLimit,
		LeiUsed:    pb.LeiUsed,
		Validators: pb.Validators,
		Proof:      ProofFromPb(pb.Proof),
		PowInfo:    nil,
	}, nil
}

type Proof struct {
	BlockHash Hash
	Height    BlockNum
	VrfProof  []byte
}

func (p *Proof) ToPb() *goproto.Proof {
	return &goproto.Proof{
		BlockHash: p.BlockHash.Bytes(),
		Height:    uint64(p.Height),
		VrfProof:  p.VrfProof,
	}
}

func ProofFromPb(pb *goproto.Proof) *Proof {
	return &Proof{
		BlockHash: BytesToHash(pb.BlockHash),
		Height:    BlockNum(pb.Height),
		VrfProof:  pb.VrfProof,
	}
}
