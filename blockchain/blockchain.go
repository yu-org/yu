package blockchain

import (
	"time"
	. "yu/common"
	"yu/config"
	ysql "yu/storage/sql"
)

type BlockChain struct {
	chain         ysql.SqlDB
	blocksFromP2p ysql.SqlDB
}

type BlocksScheme struct {
	Hash       string `gorm:"primaryKey"`
	PrevHash   string
	Height     BlockNum
	TxnRoot    string
	StateRoot  string
	Nonce      int64
	Timestamp  int64
	TxnsHashes string

	Finalize bool
}

func (BlocksScheme) TableName() string {
	return "blockchain"
}

func toBlocksScheme(b IBlock) (BlocksScheme, error) {
	header := b.GetHeader()
	bs := BlocksScheme{
		Hash:       header.GetHash().String(),
		PrevHash:   header.GetPrevHash().String(),
		Height:     header.GetHeight(),
		TxnRoot:    header.GetTxnRoot().String(),
		StateRoot:  header.GetStateRoot().String(),
		Nonce:      header.(*Header).GetNonce(),
		Timestamp:  header.GetTimestamp(),
		TxnsHashes: HashesToHex(b.GetTxnsHashes()),
		Finalize:   false,
	}
	return bs, nil
}

func (b *BlocksScheme) toBlock() (IBlock, error) {
	header := &Header{
		PrevHash:  HexToHash(b.PrevHash),
		Hash:      HexToHash(b.Hash),
		Height:    b.Height,
		TxnRoot:   HexToHash(b.TxnRoot),
		StateRoot: HexToHash(b.StateRoot),
		Nonce:     b.Nonce,
		Timestamp: b.Timestamp,
	}
	block := &Block{
		Header:     header,
		TxnsHashes: HexToHashes(b.TxnsHashes),
	}

	return block, nil
}

func NewBlockChain(cfg *config.BlockchainConf) (*BlockChain, error) {
	chain, err := ysql.NewSqlDB(&cfg.ChainDB)
	if err != nil {
		return nil, err
	}
	blocksFromP2pDB, err := ysql.NewSqlDB(&cfg.BlocksFromP2pDB)
	if err != nil {
		return nil, err
	}

	err = chain.CreateIfNotExist(&BlocksScheme{})
	if err != nil {
		return nil, err
	}

	err = blocksFromP2pDB.CreateIfNotExist(BlocksFromP2pScheme{})
	if err != nil {
		return nil, err
	}

	return &BlockChain{
		chain:         chain,
		blocksFromP2p: blocksFromP2pDB,
	}, nil
}

func (bc *BlockChain) NewDefaultBlock() IBlock {
	header := &Header{
		Timestamp: time.Now().UnixNano(),
	}
	return &Block{
		Header: header,
	}
}

func (bc *BlockChain) NewEmptyBlock() IBlock {
	return &Block{}
}

func (bc *BlockChain) SetGenesis(b IBlock) error {
	var blocks []BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Height: 0,
	}).Find(&blocks)

	if len(blocks) == 0 {
		return bc.AppendBlock(b)
	}
	return nil
}

// pending a block from other BlockChain-node for validating
func (bc *BlockChain) InsertBlockFromP2P(b IBlock) error {
	bs, err := toBlocksFromP2pScheme(b)
	if err != nil {
		return err
	}
	bc.blocksFromP2p.Db().Create(&bs)
	return nil
}

func (bc *BlockChain) GetBlocksFromP2P(height BlockNum) ([]IBlock, error) {
	var bss []BlocksFromP2pScheme
	bc.blocksFromP2p.Db().Where(&BlocksFromP2pScheme{
		Height: height,
	}).Find(&bss)
	blocks := make([]IBlock, 0)
	for _, bs := range bss {
		b, err := bs.toBlock(&Block{})
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (bc *BlockChain) FlushBlocksFromP2P(height BlockNum) error {
	bc.blocksFromP2p.Db().Where("height <= ?", height).Delete(&BlocksFromP2pScheme{})
	return nil
}

func (bc *BlockChain) AppendBlock(b IBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}
	bc.chain.Db().Create(&bs)
	return nil
}

func (bc *BlockChain) GetBlock(blockHash Hash) (IBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).First(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) Children(prevBlockHash Hash) ([]IBlock, error) {
	rows, err := bc.chain.Db().Where(&BlocksScheme{
		PrevHash: prevBlockHash.String(),
	}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []IBlock
	for rows.Next() {
		var bs BlocksScheme
		err = bc.chain.Db().ScanRows(rows, &bs)
		if err != nil {
			return nil, err
		}
		block, err := bs.toBlock()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (bc *BlockChain) Finalize(blockHash Hash) error {
	bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Update("finalize", true)
	return nil
}

func (bc *BlockChain) LastFinalized() (IBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Last(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) AllBlocks() ([]IBlock, error) {
	rows, err := bc.chain.Db().Model(&BlocksScheme{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []IBlock
	for rows.Next() {
		var bs BlocksScheme
		err = bc.chain.Db().ScanRows(rows, &bs)
		if err != nil {
			return nil, err
		}
		block, err := bs.toBlock()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (bc *BlockChain) Longest() ([]IChainStruct, error) {
	blocks, err := bc.AllBlocks()
	if err != nil {
		return nil, err
	}
	return MakeLongestChain(blocks), nil
}

func (bc *BlockChain) Heaviest() ([]IChainStruct, error) {
	blocks, err := bc.AllBlocks()
	if err != nil {
		return nil, err
	}
	return MakeHeaviestChain(blocks), nil
}

type BlocksFromP2pScheme struct {
	BlockHash    string `gorm:"primaryKey"`
	Height       BlockNum
	BlockContent string
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
		BlockHash:    b.GetHeader().GetHash().String(),
		Height:       b.GetHeader().GetHeight(),
		BlockContent: ToHex(byt),
	}, nil
}

func (bs BlocksFromP2pScheme) toBlock(b IBlock) (IBlock, error) {
	byt := FromHex(bs.BlockContent)
	return b.Decode(byt)
}
