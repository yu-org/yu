package blockchain

import (
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	ysql "github.com/yu-org/yu/storage/sql"
	"github.com/yu-org/yu/types"
	. "github.com/yu-org/yu/utils/codec"
)

type BlockChain struct {
	chain         ysql.SqlDB
	blocksFromP2p ysql.SqlDB
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

func (bc *BlockChain) ConvergeType() types.ConvergeType {
	return types.Longest
}

func (bc *BlockChain) NewEmptyBlock() types.IBlock {
	return &types.CompactBlock{Header: &types.Header{}}
}

func (bc *BlockChain) EncodeBlocks(blocks []types.IBlock) ([]byte, error) {
	var bs []*types.CompactBlock
	for _, b := range blocks {
		bs = append(bs, b.(*types.CompactBlock))
	}
	return GlobalCodec.EncodeToBytes(bs)
}

func (bc *BlockChain) DecodeBlocks(data []byte) ([]types.IBlock, error) {
	var bs []*types.CompactBlock
	err := GlobalCodec.DecodeBytes(data, &bs)
	if err != nil {
		return nil, err
	}
	var blocks []types.IBlock
	for _, b := range bs {
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (bc *BlockChain) GetGenesis() (types.IBlock, error) {
	var block BlocksScheme
	bc.chain.Db().Where("height = ?", 0).First(&block)
	return block.toBlock()
}

func (bc *BlockChain) SetGenesis(b types.IBlock) error {
	var blocks []BlocksScheme
	bc.chain.Db().Where("height = ?", 0).Find(&blocks)

	if len(blocks) == 0 {
		return bc.AppendBlock(b)
	}
	return nil
}

// pending a block from other BlockChain-node for validating
func (bc *BlockChain) InsertBlockFromP2P(b types.IBlock) error {
	if bc.ExistsBlock(b.GetHash()) {
		return nil
	}
	bs, err := toBlocksFromP2pScheme(b)
	if err != nil {
		return err
	}
	bc.blocksFromP2p.Db().Create(&bs)
	return nil
}

func (bc *BlockChain) TakeP2pBlocksBefore(height BlockNum) (map[BlockNum][]types.IBlock, error) {
	var bsp []BlocksFromP2pScheme
	bc.blocksFromP2p.Db().Where("height < ?", height).Order("height").Find(&bsp)
	blocks := bspToBlocks(bsp)
	hBlocks := make(map[BlockNum][]types.IBlock, 0)
	for _, block := range blocks {
		height := block.GetHeight()
		hBlocks[height] = append(hBlocks[height], block)
	}

	for _, b := range blocks {
		bc.blocksFromP2p.Db().Delete(&BlocksFromP2pScheme{BlockHash: b.GetHash().String()})
	}
	return hBlocks, nil
}

func (bc *BlockChain) TakeP2pBlocks(height BlockNum) ([]types.IBlock, error) {
	var bsp []BlocksFromP2pScheme
	bc.blocksFromP2p.Db().Where("height = ?", height).Find(&bsp)

	for _, bs := range bsp {
		bc.blocksFromP2p.Db().Delete(&BlocksFromP2pScheme{BlockHash: bs.BlockHash})
	}
	return bspToBlocks(bsp), nil
}

func (bc *BlockChain) AppendBlock(b types.IBlock) error {
	if bc.ExistsBlock(b.GetHash()) {
		return nil
	}
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

func (bc *BlockChain) GetBlock(blockHash Hash) (types.IBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).First(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) UpdateBlock(b types.IBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	bc.chain.Db().Where(&BlocksScheme{
		Hash: b.GetHash().String(),
	}).Updates(bs)
	return nil
}

func (bc *BlockChain) Children(prevBlockHash Hash) ([]types.IBlock, error) {
	rows, err := bc.chain.Db().Where(&BlocksScheme{
		PrevHash: prevBlockHash.String(),
	}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []types.IBlock
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

func (bc *BlockChain) LastFinalized() (types.IBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Last(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) GetEndBlock() (types.IBlock, error) {
	var bs BlocksScheme
	bc.chain.Db().Raw("select * from blockchain where height = (select max(height) from blockchain)").First(&bs)
	return bs.toBlock()
}

func (bc *BlockChain) GetAllBlocks() ([]types.IBlock, error) {
	rows, err := bc.chain.Db().Model(&BlocksScheme{}).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []types.IBlock
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

func (bc *BlockChain) GetRangeBlocks(startHeight, endHeight BlockNum) ([]types.IBlock, error) {
	var bss []BlocksScheme
	bc.chain.Db().Where("height BETWEEN ? AND ?", startHeight, endHeight).Find(&bss)
	return bssToBlocks(bss), nil
}

func (bc *BlockChain) Chain() (types.IChainStruct, error) {
	return bc.LongestChain()
}

func (bc *BlockChain) LongestChain() (types.IChainStruct, error) {
	block, err := bc.GetEndBlock()
	if err != nil {
		return nil, err
	}
	prevHash := block.GetPrevHash()
	chain := NewEmptyChain(block)
	for block.GetHeight() > 0 {
		prevBlock, err := bc.GetBlock(prevHash)
		if err != nil {
			return nil, err
		}
		chain.InsertPrev(prevBlock)
		block = prevBlock
	}
	return chain, nil
}

func (bc *BlockChain) HeaviestChains() ([]types.IChainStruct, error) {
	blocks, err := bc.GetAllBlocks()
	if err != nil {
		return nil, err
	}
	return MakeHeaviestChain(blocks), nil
}

func (bc *BlockChain) FinalizedChain() (types.IChainStruct, error) {
	var bss []BlocksScheme
	bc.chain.Db().Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Find(&bss)
	blocks := bssToBlocks(bss)
	return MakeFinalizedChain(blocks), nil
}
