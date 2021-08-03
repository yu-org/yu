package blockchain

import (
	. "github.com/Lawliet-Chan/yu/common"
	"github.com/libp2p/go-libp2p-core/peer"
)

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

	LeiLimit uint64
	LeiUsed  uint64

	Length   uint64
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
		PeerID:     b.GetPeerID().String(),

		LeiLimit: b.GetLeiLimit(),
		LeiUsed:  b.GetLeiUsed(),

		Length:   b.(*Block).ChainLength,
		Finalize: false,
	}
	return bs, nil
}

func (b *BlocksScheme) toBlock() (IBlock, error) {
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

	header := &Header{
		PrevHash:  HexToHash(b.PrevHash),
		Hash:      HexToHash(b.Hash),
		Height:    b.Height,
		TxnRoot:   HexToHash(b.TxnRoot),
		StateRoot: HexToHash(b.StateRoot),
		Nonce:     b.Nonce,
		Timestamp: b.Timestamp,
		PeerID:    PeerID,
		LeiLimit:  b.LeiLimit,
		LeiUsed:   b.LeiUsed,
	}
	block := &Block{
		Header:      header,
		TxnsHashes:  HexToHashes(b.TxnsHashes),
		ChainLength: b.Length,
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
