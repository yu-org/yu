package blockchain

import (
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/types"
	"github.com/yu-org/yu/types/goproto"
)

type BlocksScheme struct {
	Hash       string `gorm:"primaryKey"`
	PrevHash   string
	Height     BlockNum
	TxnRoot    string
	StateRoot  string
	Timestamp  uint64
	TxnsHashes string
	PeerID     string

	LeiLimit uint64
	LeiUsed  uint64

	Finalize bool

	Pubkey    string
	Signature string

	Validators []byte

	Nonce      uint64
	Difficulty uint64

	ProofBlock  string
	ProofHeight BlockNum
	VrfProof    string

	Extra string
}

func (BlocksScheme) TableName() string {
	return "blockchain"
}

func toBlocksScheme(b *CompactBlock) (BlocksScheme, error) {
	validators, err := proto.Marshal(b.Validators)
	if err != nil {
		return BlocksScheme{}, err
	}
	return BlocksScheme{
		Hash:       b.Hash.String(),
		PrevHash:   b.PrevHash.String(),
		Height:     b.Height,
		TxnRoot:    b.TxnRoot.String(),
		StateRoot:  b.StateRoot.String(),
		Timestamp:  b.Timestamp,
		TxnsHashes: HashesToHex(b.TxnsHashes),
		PeerID:     b.PeerID.String(),
		Finalize:   false,
		LeiLimit:   b.LeiLimit,
		LeiUsed:    b.LeiUsed,

		Pubkey:    ToHex(b.Pubkey),
		Signature: ToHex(b.Signature),

		Validators: validators,

		Nonce:      b.Nonce,
		Difficulty: b.Difficulty,

		ProofBlock:  b.ProofBlockHash.String(),
		ProofHeight: b.ProofHeight,
		VrfProof:    ToHex(b.Proof),

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
		PrevHash:  HexToHash(b.PrevHash),
		Hash:      HexToHash(b.Hash),
		Height:    b.Height,
		TxnRoot:   HexToHash(b.TxnRoot),
		StateRoot: HexToHash(b.StateRoot),
		Timestamp: b.Timestamp,
		PeerID:    PeerID,

		LeiLimit: b.LeiLimit,
		LeiUsed:  b.LeiUsed,

		Pubkey:    FromHex(b.Pubkey),
		Signature: FromHex(b.Signature),

		Validators: &validators,
		Nonce:      b.Nonce,
		Difficulty: b.Difficulty,

		ProofBlockHash: HexToHash(b.ProofBlock),
		ProofHeight:    b.ProofHeight,
		Proof:          FromHex(b.VrfProof),

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
//	Pubkey        string
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
//		Pubkey:        pb.String(),
//		ProposeWeight: pb.ProposeWeight,
//		VoteWeight:    pb.VoteWeight,
//	}
//}
//
//func (v *ValidatorScheme) toValidator() *goproto.Validator {
//	return &goproto.Validator{
//		PubKey:        FromHex(v.Pubkey),
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
