package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types/goproto"
	"github.com/yu-org/yu/infra/trie"
)

type Block struct {
	*Header
	Txns SignedTxns
}

func (b *Block) Compact() *CompactBlock {
	return &CompactBlock{
		Header:     b.Header,
		TxnsHashes: b.Txns.Hashes(),
	}
}

func (b *Block) CopyFrom(other *Block) {
	*b = *other
}

func (b *Block) UseLei(lei uint64) {
	b.Header.LeiUsed += lei
}

func (b *Block) SetTxns(txns SignedTxns) {
	b.Txns = txns
}

func (b *Block) Encode() ([]byte, error) {
	return proto.Marshal(b.ToPb())
}

func DecodeBlock(data []byte) (*Block, error) {
	var b goproto.Block
	err := proto.Unmarshal(data, &b)
	if err != nil {
		return nil, err
	}
	return BlockFromPb(&b)
}

func EncodeBlocks(blocks []*Block) ([]byte, error) {
	bs := make([]*goproto.Block, 0)
	for _, cb := range blocks {
		bs = append(bs, cb.ToPb())
	}
	pb := &goproto.Blocks{
		Blocks: bs,
	}
	return proto.Marshal(pb)
}

func DecodeBlocks(data []byte) ([]*Block, error) {
	var bs goproto.Blocks
	err := proto.Unmarshal(data, &bs)
	if err != nil {
		return nil, err
	}
	blocks := make([]*Block, 0)
	for _, b := range bs.Blocks {
		block, err := BlockFromPb(b)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (b *Block) ToPb() *goproto.Block {
	return &goproto.Block{
		Header: b.Header.ToPb(),
		Txns:   b.Txns.ToPb(),
	}
}

func BlockFromPb(pb *goproto.Block) (*Block, error) {
	txns, err := SignedTxnsFromPb(pb.Txns)
	if err != nil {
		return nil, err
	}
	return &Block{
		Header: HeaderFromPb(pb.Header),
		Txns:   txns,
	}, nil
}

type CompactBlock struct {
	*Header
	TxnsHashes []Hash
}

func EncodeCompactBlocks(blocks []*CompactBlock) ([]byte, error) {
	cbs := make([]*goproto.CompactBlock, 0)
	for _, cb := range blocks {
		cbs = append(cbs, cb.ToPb())
	}
	pb := &goproto.CompactBlocks{
		Blocks: cbs,
	}
	return proto.Marshal(pb)
}

func DecodeCompactBlocks(data []byte) ([]*CompactBlock, error) {
	var bs goproto.CompactBlocks
	err := proto.Unmarshal(data, &bs)
	if err != nil {
		return nil, err
	}
	cbs := make([]*CompactBlock, 0)
	for _, b := range bs.Blocks {
		cbs = append(cbs, CompactBlockFromPb(b))
	}
	return cbs, nil
}

func (b *CompactBlock) CopyFrom(other *CompactBlock) {
	*b = *other
}

func (cb *CompactBlock) Encode() ([]byte, error) {
	return proto.Marshal(cb.ToPb())
}

func DecodeCompactBlock(byt []byte) (*CompactBlock, error) {
	var b goproto.CompactBlock
	err := proto.Unmarshal(byt, &b)
	if err != nil {
		return nil, err
	}
	return CompactBlockFromPb(&b), nil
}

func (cb *CompactBlock) ToPb() *goproto.CompactBlock {
	return &goproto.CompactBlock{
		Header:     cb.Header.ToPb(),
		TxnsHashes: HashesToTwoBytes(cb.TxnsHashes),
	}
}

func CompactBlockFromPb(pb *goproto.CompactBlock) *CompactBlock {
	header := HeaderFromPb(pb.Header)
	return &CompactBlock{
		Header:     header,
		TxnsHashes: TwoBytesToHashes(pb.TxnsHashes),
	}
}

func IfLeiOut(Lei uint64, block *Block) bool {
	return Lei+block.LeiUsed > block.LeiLimit
}

func MakeTxnRoot(txns []*SignedTxn) (Hash, error) {
	txnsBytes := make([]Hash, 0)
	for _, tx := range txns {
		hash := tx.TxnHash
		txnsBytes = append(txnsBytes, hash)
	}
	mTree := trie.NewMerkleTree(txnsBytes)
	return mTree.RootNode.Data, nil
}

type Header struct {
	ChainID  uint64
	PrevHash Hash
	Hash     Hash
	Height   BlockNum

	TxnRoot     Hash
	StateRoot   Hash
	ReceiptRoot Hash

	Timestamp uint64
	PeerID    peer.ID

	Extra []byte

	LeiLimit uint64
	LeiUsed  uint64

	MinerPubkey    []byte
	MinerSignature []byte

	Validators     []*Validator
	ProofBlockHash Hash
	ProofHeight    BlockNum
	Proof          []byte

	Nonce      uint64
	Difficulty uint64
}

func (h *Header) ToPb() *goproto.Header {
	return &goproto.Header{
		ChainId:     h.ChainID,
		Hash:        h.Hash.Bytes(),
		PrevHash:    h.PrevHash.Bytes(),
		Height:      uint64(h.Height),
		TxnRoot:     h.TxnRoot.Bytes(),
		StateRoot:   h.StateRoot.Bytes(),
		ReceiptRoot: h.ReceiptRoot.Bytes(),
		Timestamp:   h.Timestamp,
		PeerId:      h.PeerID.String(),
		LeiLimit:    h.LeiLimit,
		LeiUsed:     h.LeiUsed,
		Validators:  ValidatorsToPb(h.Validators),

		MinerPubkey:    h.MinerPubkey,
		MinerSignature: h.MinerSignature,

		ProofBlockHash: h.ProofBlockHash.Bytes(),
		ProofHeight:    uint64(h.ProofHeight),
		Proof:          h.Proof,

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
	if pb.PeerId == "" {
		peerID = peer.ID("")
	} else {
		peerID, err = peer.Decode(pb.PeerId)
		if err != nil {
			logrus.Panicf("peerID(%s) decode error: %v", pb.PeerId, err)
		}
	}

	return &Header{
		ChainID:     pb.ChainId,
		PrevHash:    BytesToHash(pb.PrevHash),
		Hash:        BytesToHash(pb.Hash),
		Height:      BlockNum(pb.Height),
		TxnRoot:     BytesToHash(pb.TxnRoot),
		StateRoot:   BytesToHash(pb.StateRoot),
		ReceiptRoot: BytesToHash(pb.ReceiptRoot),

		Timestamp: pb.Timestamp,
		PeerID:    peerID,

		LeiLimit:   pb.LeiLimit,
		LeiUsed:    pb.LeiUsed,
		Validators: ValidatorsFromPb(pb.Validators),

		MinerPubkey:    pb.MinerPubkey,
		MinerSignature: pb.MinerSignature,

		ProofBlockHash: BytesToHash(pb.ProofBlockHash),
		ProofHeight:    BlockNum(pb.ProofHeight),
		Proof:          pb.Proof,

		Nonce:      pb.Nonce,
		Difficulty: pb.Difficulty,

		Extra: pb.Extra,
	}
}

type Validator struct {
	PubKey        []byte
	ProposeWeight uint64
	VoteWeight    uint64
}

func ValidatorsToPb(vs []*Validator) *goproto.Validators {
	validators := make([]*goproto.Validator, 0)
	for _, v := range vs {
		validators = append(validators, &goproto.Validator{
			PubKey:        v.PubKey,
			ProposeWeight: v.ProposeWeight,
			VoteWeight:    v.VoteWeight,
		})
	}
	return &goproto.Validators{Validators: validators}
}

func ValidatorsFromPb(vs *goproto.Validators) []*Validator {
	validators := make([]*Validator, 0)
	for _, v := range vs.Validators {
		validators = append(validators, &Validator{
			PubKey:        v.PubKey,
			ProposeWeight: v.ProposeWeight,
			VoteWeight:    v.VoteWeight,
		})
	}
	return validators
}
