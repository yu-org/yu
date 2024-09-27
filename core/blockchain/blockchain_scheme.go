package blockchain

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/core/types/goproto"
)

type BlocksScheme struct {
	ChainID     uint64
	Hash        string `gorm:"primaryKey;type:varchar(100)"`
	PrevHash    string
	Height      BlockNum `gorm:"index"`
	TxnRoot     string
	StateRoot   string
	ReceiptRoot string

	Timestamp  uint64
	TxnsHashes string
	PeerID     string

	LeiLimit uint64
	LeiUsed  uint64

	Finalize bool

	MinerPubkey    string
	MinerSignature string

	Validators []byte

	Nonce      uint64
	Difficulty uint64

	ProofBlock  string
	ProofHeight BlockNum
	Proof       string

	Extra string
}

func (BlocksScheme) TableName() string {
	return "blockchain"
}

func toBlocksScheme(b *CompactBlock) (BlocksScheme, error) {
	validators, err := proto.Marshal(ValidatorsToPb(b.Validators))
	if err != nil {
		return BlocksScheme{}, err
	}
	return BlocksScheme{
		ChainID:     b.ChainID,
		Hash:        b.Hash.String(),
		PrevHash:    b.PrevHash.String(),
		Height:      b.Height,
		TxnRoot:     b.TxnRoot.String(),
		StateRoot:   b.StateRoot.String(),
		ReceiptRoot: b.ReceiptRoot.String(),
		Timestamp:   b.Timestamp,
		TxnsHashes:  HashesToHex(b.TxnsHashes),
		PeerID:      b.PeerID.String(),
		Finalize:    false,
		LeiLimit:    b.LeiLimit,
		LeiUsed:     b.LeiUsed,

		MinerPubkey:    ToHex(b.MinerPubkey),
		MinerSignature: ToHex(b.MinerSignature),

		Validators: validators,

		Nonce:      b.Nonce,
		Difficulty: b.Difficulty,

		ProofBlock:  b.ProofBlockHash.String(),
		ProofHeight: b.ProofHeight,
		Proof:       ToHex(b.Proof),

		Extra: string(b.Extra),
	}, nil
}

func (b *BlocksScheme) toBlock() (*CompactBlock, error) {
	var (
		PeerID peer.ID
		err    error
	)
	if b.PeerID == "" {
		PeerID = peer.ID("")
	} else {
		PeerID, err = peer.Decode(b.PeerID)
		if err != nil {
			return nil, err
		}
	}

	var validators goproto.Validators
	err = proto.Unmarshal(b.Validators, &validators)
	if err != nil {
		return nil, err
	}

	header := &Header{
		ChainID:     b.ChainID,
		PrevHash:    HexToHash(b.PrevHash),
		Hash:        HexToHash(b.Hash),
		Height:      b.Height,
		TxnRoot:     HexToHash(b.TxnRoot),
		StateRoot:   HexToHash(b.StateRoot),
		ReceiptRoot: HexToHash(b.ReceiptRoot),
		Timestamp:   b.Timestamp,
		PeerID:      PeerID,

		LeiLimit: b.LeiLimit,
		LeiUsed:  b.LeiUsed,

		MinerPubkey:    FromHex(b.MinerPubkey),
		MinerSignature: FromHex(b.MinerSignature),

		Validators: ValidatorsFromPb(&validators),
		Nonce:      b.Nonce,
		Difficulty: b.Difficulty,

		ProofBlockHash: HexToHash(b.ProofBlock),
		ProofHeight:    b.ProofHeight,
		Proof:          FromHex(b.Proof),

		Extra: []byte(b.Extra),
	}
	block := &CompactBlock{
		Header:     header,
		TxnsHashes: HexToHashes(b.TxnsHashes),
	}

	return block, nil
}

func bssToBlocks(bss []BlocksScheme) []*CompactBlock {
	blocks := make([]*CompactBlock, 0)
	for _, bs := range bss {
		b, err := bs.toBlock()
		if err != nil {
			logrus.Panic("blockscheme to CompactBlock error: ", err)
		}
		blocks = append(blocks, b)
	}
	return blocks
}

//type ValidatorScheme struct {
//	BlockHash     string
//	MinerPubkey        string
//	ProposeWeight uint64
//	VoteWeight    uint64
//}
//
//func (ValidatorScheme) TableName() string {
//	return "blockchain"
//}
//
//func toValidatorScheme(pb *goproto.Validator) ValidatorScheme {
//	return ValidatorScheme{
//		BlockHash:     "",
//		MinerPubkey:        pb.String(),
//		ProposeWeight: pb.ProposeWeight,
//		VoteWeight:    pb.VoteWeight,
//	}
//}
//
//func (v *ValidatorScheme) toValidator() *goproto.Validator {
//	return &goproto.Validator{
//		PubKey:        FromHex(v.MinerPubkey),
//		ProposeWeight: v.ProposeWeight,
//		VoteWeight:    v.VoteWeight,
//	}
//}
//
//func vssToValidators(vss []ValidatorScheme) []*goproto.Validator {
//	validators := make([]*goproto.Validator, 0)
//	for _, vs := range vss {
//		validators = append(validators, vs.toValidator())
//	}
//	return validators
//}
