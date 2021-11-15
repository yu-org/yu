package blockchain

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	ysql "github.com/yu-org/yu/storage/sql"
	. "github.com/yu-org/yu/types"
	. "github.com/yu-org/yu/utils/codec"
)

type BlockChain struct {
	chain ysql.SqlDB
}

func NewBlockChain(cfg *config.BlockchainConf) (*BlockChain, error) {
	chain, err := ysql.NewSqlDB(&cfg.ChainDB)
	if err != nil {
		return nil, err
	}

	err = chain.CreateIfNotExist(&BlocksScheme{})
	if err != nil {
		return nil, err
	}

	return &BlockChain{
		chain: chain,
	}, nil
}

func (bc *BlockChain) ConvergeType() ConvergeType {
	return Longest
}

func (bc *BlockChain) NewEmptyBlock() *CompactBlock {
	return &CompactBlock{Header: &Header{}}
}

func (bc *BlockChain) EncodeBlocks(blocks []*CompactBlock) ([]byte, error) {
	var bs []*CompactBlock
	for _, b := range blocks {
		bs = append(bs, b)
	}
	return GlobalCodec.EncodeToBytes(bs)
}

func (bc *BlockChain) DecodeBlocks(data []byte) ([]*CompactBlock, error) {
	var bs []*CompactBlock
	err := GlobalCodec.DecodeBytes(data, &bs)
	if err != nil {
		return nil, err
	}
	var blocks []*CompactBlock
	for _, b := range bs {
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (bc *BlockChain) GetGenesis() (*CompactBlock, error) {
	var block BlocksScheme
	bc.chain.Db().Where("height = ?", 0).First(&block)
	return block.toBlock()
}

func (bc *BlockChain) SetGenesis(b *CompactBlock) error {
	var blocks []BlocksScheme
	bc.chain.Db().Where("height = ?", 0).Find(&blocks)

	if len(blocks) == 0 {
		return bc.AppendBlock(b)
	}
	return nil
}

func (bc *BlockChain) AppendBlock(b *CompactBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	return bc.chain.Db().Create(bs).Error
}

func (bc *BlockChain) ExistsBlock(blockHash Hash) bool {
	var bss []BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Find(&bss)

	return len(bss) > 0
}

func (bc *BlockChain) GetBlock(blockHash Hash) (*CompactBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).First(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) UpdateBlock(b *CompactBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	bc.chain.Db().Where(&BlocksScheme{
		Hash: b.Hash.String(),
	}).Updates(bs)
	return nil
}

func (bc *BlockChain) Children(prevBlockHash Hash) ([]*CompactBlock, error) {
	rows, err := bc.chain.Db().Where(&BlocksScheme{
		PrevHash: prevBlockHash.String(),
	}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*CompactBlock
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
	}).Updates(BlocksScheme{Finalize: true})
	return nil
}

func (bc *BlockChain) LastFinalized() (*CompactBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Last(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) GetEndBlock() (*CompactBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Raw("select * from blockchain where height = (select max(height) from blockchain)").First(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) GetAllBlocks() ([]*CompactBlock, error) {
	rows, err := bc.chain.Db().Model(&BlocksScheme{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*CompactBlock
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

func (bc *BlockChain) GetRangeBlocks(startHeight, endHeight BlockNum) ([]*CompactBlock, error) {
	var bss []BlocksScheme
	bc.chain.Db().Where("height BETWEEN ? AND ?", startHeight, endHeight).Find(&bss)
	return bssToBlocks(bss), nil
}

func (bc *BlockChain) Chain() (IChainStruct, error) {
	return bc.LongestChain()
}

func (bc *BlockChain) LongestChain() (IChainStruct, error) {
	block, err := bc.GetEndBlock()
	if err != nil {
		return nil, err
	}
	prevHash := block.PrevHash
	chain := NewEmptyChain(block)
	for block.Height > 0 {
		prevBlock, err := bc.GetBlock(prevHash)
		if err != nil {
			return nil, err
		}
		chain.InsertPrev(prevBlock)
		block = prevBlock
	}
	return chain, nil
}

func (bc *BlockChain) HeaviestChains() ([]IChainStruct, error) {
	blocks, err := bc.GetAllBlocks()
	if err != nil {
		return nil, err
	}
	return MakeHeaviestChain(blocks), nil
}

func (bc *BlockChain) FinalizedChain() (IChainStruct, error) {
	var bss []BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Find(&bss)
	blocks := bssToBlocks(bss)
	return MakeFinalizedChain(blocks), nil
}
