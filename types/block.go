package types

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/trie"
	"github.com/yu-org/yu/types/goproto"
	. "github.com/yu-org/yu/yerror"
)

type Block struct {
	*CompactBlock
	Txns SignedTxns
}

type CompactBlock struct {
	Header     *Header
	TxnsHashes []Hash
}

func IfLeiOut(Lei uint64, block IBlock) bool {
	return Lei+block.GetLeiUsed() > block.GetLeiLimit()
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

	Validators []*goproto.Validator
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
	Pubkey    []byte
	Signature []byte
	VrfProof  []byte
}

func (p *Proof) ToPb() *goproto.Proof {
	return &goproto.Proof{
		BlockHash: p.BlockHash.Bytes(),
		Height:    uint64(p.Height),
		PubKey:    p.Pubkey,
		Signature: p.Signature,
		VrfProof:  p.VrfProof,
	}
}

func ProofFromPb(pb *goproto.Proof) *Proof {
	return &Proof{
		BlockHash: BytesToHash(pb.BlockHash),
		Height:    BlockNum(pb.Height),
		Pubkey:    pb.PubKey,
		Signature: pb.Signature,
		VrfProof:  pb.VrfProof,
	}
}
