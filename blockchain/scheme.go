package blockchain

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/Lawliet-Chan/yu/keypair"
	. "github.com/Lawliet-Chan/yu/result"
	. "github.com/Lawliet-Chan/yu/txn"
	"github.com/libp2p/go-libp2p-core/peer"
	"gorm.io/gorm"
)

type TxnScheme struct {
	TxnHash   string `gorm:"primaryKey"`
	Pubkey    string
	KeyType   string
	Signature string
	RawTxn    string

	BlockHash string
}

func (TxnScheme) TableName() string {
	return "txns"
}

func newTxnScheme(blockHash Hash, stxn *SignedTxn) (TxnScheme, error) {
	txnSm, err := toTxnScheme(stxn)
	if err != nil {
		return TxnScheme{}, err
	}
	txnSm.BlockHash = blockHash.String()
	return txnSm, nil
}

func toTxnScheme(stxn *SignedTxn) (TxnScheme, error) {
	rawTxnByt, err := stxn.GetRaw().Encode()
	if err != nil {
		return TxnScheme{}, err
	}
	return TxnScheme{
		TxnHash:   stxn.GetTxnHash().String(),
		Pubkey:    stxn.GetPubkey().String(),
		KeyType:   stxn.GetPubkey().Type(),
		Signature: ToHex(stxn.GetSignature()),
		RawTxn:    ToHex(rawTxnByt),
		BlockHash: "",
	}, nil
}

func (t TxnScheme) toTxn() (*SignedTxn, error) {
	ut := &UnsignedTxn{}
	rawTxn, err := ut.Decode(FromHex(t.RawTxn))
	if err != nil {
		return nil, err
	}
	pubkey, err := keypair.PubKeyFromBytes(t.KeyType, FromHex(t.Pubkey))
	if err != nil {
		return nil, err
	}
	return &SignedTxn{
		Raw:       rawTxn,
		TxnHash:   HexToHash(t.TxnHash),
		Pubkey:    pubkey,
		Signature: FromHex(t.Signature),
	}, nil
}

type EventScheme struct {
	gorm.Model
	Caller     string
	BlockStage string
	BlockHash  string
	Height     BlockNum
	TripodName string
	ExecName   string
	Value      string
}

func (EventScheme) TableName() string {
	return "events"
}

func toEventScheme(event *Event) (EventScheme, error) {
	return EventScheme{
		Caller:     event.Caller.String(),
		BlockStage: event.BlockStage,
		BlockHash:  event.BlockHash.String(),
		Height:     event.Height,
		TripodName: event.TripodName,
		ExecName:   event.ExecName,
		Value:      event.Value,
	}, nil
}

func (e EventScheme) toEvent() (*Event, error) {
	return &Event{
		Caller:     HexToAddress(e.Caller),
		BlockStage: e.BlockStage,
		BlockHash:  HexToHash(e.BlockHash),
		Height:     e.Height,
		TripodName: e.TripodName,
		ExecName:   e.ExecName,
		Value:      e.Value,
	}, nil

}

type ErrorScheme struct {
	gorm.Model
	Caller     string
	BlockStage string
	BlockHash  string
	Height     BlockNum
	TripodName string
	ExecName   string
	Error      string
}

func (ErrorScheme) TableName() string {
	return "errors"
}

func toErrorScheme(err *Error) ErrorScheme {
	return ErrorScheme{
		Caller:     err.Caller.String(),
		BlockStage: err.BlockStage,
		BlockHash:  err.BlockHash.String(),
		Height:     err.Height,
		TripodName: err.TripodName,
		ExecName:   err.ExecName,
		Error:      err.Err,
	}
}

func (e ErrorScheme) toError() *Error {
	return &Error{
		Caller:     HexToAddress(e.Caller),
		BlockStage: e.BlockStage,
		BlockHash:  HexToHash(e.BlockHash),
		Height:     e.Height,
		TripodName: e.TripodName,
		ExecName:   e.ExecName,
		Err:        e.Error,
	}
}

type BlocksScheme struct {
	Hash       string `gorm:"primaryKey"`
	PrevHash   string
	Height     BlockNum
	TxnRoot    string
	StateRoot  string
	Nonce      uint64
	Timestamp  uint64
	TxnsHashes string
	PeerID     string

	Finalize bool
}

func (BlocksScheme) TableName() string {
	return "blockchain"
}

func toBlocksScheme(b IBlock) (BlocksScheme, error) {
	bs := BlocksScheme{
		Hash:       b.GetHash().String(),
		PrevHash:   b.GetPrevHash().String(),
		Height:     b.GetHeight(),
		TxnRoot:    b.GetTxnRoot().String(),
		StateRoot:  b.GetStateRoot().String(),
		Nonce:      b.GetHeader().(*Header).Nonce,
		Timestamp:  b.GetTimestamp(),
		TxnsHashes: HashesToHex(b.GetTxnsHashes()),
		PeerID:     b.GetProducerPeer().String(),
		Finalize:   false,
	}
	return bs, nil
}

func (b *BlocksScheme) toBlock() (IBlock, error) {
	var (
		peerID peer.ID
		err    error
	)
	if b.PeerID == "" {
		peerID = peer.ID("")
	} else {
		peerID, err = peer.Decode(b.PeerID)
		if err != nil {
			return nil, err
		}
	}

	header := &Header{
		PrevHash:     HexToHash(b.PrevHash),
		Hash:         HexToHash(b.Hash),
		Height:       b.Height,
		TxnRoot:      HexToHash(b.TxnRoot),
		StateRoot:    HexToHash(b.StateRoot),
		Nonce:        b.Nonce,
		Timestamp:    b.Timestamp,
		ProducerPeer: peerID,
	}
	block := &Block{
		Header:     header,
		TxnsHashes: HexToHashes(b.TxnsHashes),
	}

	return block, nil
}

type BlocksFromP2pScheme struct {
	BlockHash     string `gorm:"primaryKey"`
	Height        BlockNum
	BlockContent  string
	BlockProducer string
}

func (BlocksFromP2pScheme) TableName() string {
	return "blocks_from_p2p"
}

func toBlocksFromP2pScheme(b IBlock) (BlocksFromP2pScheme, error) {
	byt, err := b.Encode()
	if err != nil {
		return BlocksFromP2pScheme{}, err
	}
	return BlocksFromP2pScheme{
		BlockHash:    b.GetHash().String(),
		Height:       b.GetHeight(),
		BlockContent: ToHex(byt),
	}, nil
}

func (bs BlocksFromP2pScheme) toBlock() (IBlock, error) {
	byt := FromHex(bs.BlockContent)
	b := &Block{}
	return b.Decode(byt)
}

func bssToBlocks(bss []BlocksScheme) []IBlock {
	blocks := make([]IBlock, 0)
	for _, bs := range bss {
		b, err := bs.toBlock()
		if err != nil {
			return nil
		}
		blocks = append(blocks, b)
	}
	return blocks
}

func bspToBlocks(bsp []BlocksFromP2pScheme) []IBlock {
	blocks := make([]IBlock, 0)
	for _, bs := range bsp {
		b, err := bs.toBlock()
		if err != nil {
			return nil
		}
		blocks = append(blocks, b)
	}
	return blocks
}
