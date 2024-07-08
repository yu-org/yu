package blockchain

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	ysql "github.com/yu-org/yu/infra/storage/sql"
	"gorm.io/gorm"
)

type BlockChain struct {
	nodeType int
	chain    ysql.SqlDB
	ItxDB
}

func NewBlockChain(nodeType int, cfg *config.BlockchainConf, txdb ItxDB) *BlockChain {
	chain, err := ysql.NewSqlDB(&cfg.ChainDB)
	if err != nil {
		logrus.Fatal("init blockchain SQL db error: ", err)
	}

	err = chain.CreateIfNotExist(&BlocksScheme{})
	if err != nil {
		logrus.Fatal("create blockchain scheme: ", err)
	}

	return &BlockChain{
		nodeType: nodeType,
		chain:    chain,
		ItxDB:    txdb,
	}
}

func (bc *BlockChain) ConvergeType() ConvergeType {
	return Longest
}

func (bc *BlockChain) NewEmptyBlock() *Block {
	return &Block{Header: &Header{}, Txns: nil}
}

func (bc *BlockChain) GetGenesis() (*Block, error) {
	var block BlocksScheme
	err := bc.chain.Db().Where("height = ?", 0).First(&block).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, yerror.ErrBlockNotFound
		}
		return nil, err
	}
	cb, err := block.toBlock()
	if err != nil {
		return nil, err
	}
	b := &Block{
		Header: cb.Header,
	}
	for _, hash := range cb.TxnsHashes {
		txn, err := bc.ItxDB.GetTxn(hash)
		if err != nil {
			return nil, err
		}
		b.Txns = append(b.Txns, txn)
	}
	return b, nil
}

func (bc *BlockChain) SetGenesis(b *Block) error {
	var blocks []BlocksScheme
	err := bc.chain.Db().Where("height = ?", 0).Find(&blocks).Error
	if err != nil {
		return err
	}

	if len(blocks) == 0 {
		return bc.AppendBlock(b)
	}
	return nil
}

func (bc *BlockChain) AppendBlock(b *Block) error {
	cb := b.Compact()
	if bc.nodeType == LightNode {
		cb.TxnsHashes = nil
	}
	err := bc.appendCompactBlock(cb)
	if err != nil {
		return err
	}
	return bc.ItxDB.SetTxns(b.Txns)
}

func (bc *BlockChain) appendCompactBlock(b *CompactBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	return bc.chain.Db().Create(bs).Error
}

func (bc *BlockChain) ExistsBlock(blockHash Hash) (bool, error) {
	var bss []BlocksScheme
	err := bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Find(&bss).Error
	if err != nil {
		return false, err
	}

	return len(bss) > 0, nil
}

func (bc *BlockChain) GetCompactBlock(blockHash Hash) (*CompactBlock, error) {
	var bs BlocksScheme
	err := bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).First(&bs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, yerror.ErrBlockNotFound
		}
		return nil, err
	}
	return bs.toBlock()
}

func (bc *BlockChain) GetBlock(blockHash Hash) (*Block, error) {
	cBlock, err := bc.GetCompactBlock(blockHash)
	if err != nil {
		return nil, err
	}
	txns, err := bc.ItxDB.GetTxns(cBlock.TxnsHashes)
	if err != nil {
		return nil, err
	}
	return &Block{
		Header: cBlock.Header,
		Txns:   txns,
	}, nil
}

func (bc *BlockChain) GetCompactBlockByHeight(height BlockNum) (*CompactBlock, error) {
	var bs BlocksScheme
	err := bc.chain.Db().Where(&BlocksScheme{
		Height:   height,
		Finalize: true,
	}).First(&bs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, yerror.ErrBlockNotFound
		}
		return nil, err
	}
	return bs.toBlock()
}

func (bc *BlockChain) GetBlockByHeight(height BlockNum) (*Block, error) {
	cBlock, err := bc.GetCompactBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	txns, err := bc.ItxDB.GetTxns(cBlock.TxnsHashes)
	if err != nil {
		return nil, err
	}
	return &Block{
		Header: cBlock.Header,
		Txns:   txns,
	}, nil
}

func (bc *BlockChain) GetAllBlocksByHeight(height BlockNum) ([]*CompactBlock, error) {
	var bss []BlocksScheme
	err := bc.chain.Db().Where(&BlocksScheme{Height: height}).Find(&bss).Error
	if err != nil {
		return nil, err
	}
	return bssToBlocks(bss), nil
}

func (bc *BlockChain) UpdateBlock(b *CompactBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	return bc.chain.Db().Where(&BlocksScheme{
		Hash: b.Hash.String(),
	}).Updates(bs).Error
}

func (bc *BlockChain) UpdateBlockByHeight(b *CompactBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}

	return bc.chain.Db().Where(&BlocksScheme{
		Height: b.Height,
	}).Updates(bs).Error
}

func (bc *BlockChain) Children(prevBlockHash Hash) ([]*CompactBlock, error) {
	rows, err := bc.chain.Db().Model(&BlocksScheme{}).Where(BlocksScheme{
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
	return bc.chain.Db().Model(&BlocksScheme{}).Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Updates(BlocksScheme{Finalize: true}).Error
}

func (bc *BlockChain) LastFinalized() (*CompactBlock, error) {
	var bss []BlocksScheme
	err := bc.chain.Db().Model(&BlocksScheme{}).Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Find(&bss).Error
	if err != nil {
		return nil, err
	}
	bs := bss[len(bss)-1]
	return bs.toBlock()
}

func (bc *BlockChain) GetEndBlock() (*CompactBlock, error) {
	var bs BlocksScheme
	err := bc.chain.Db().Raw("select * from blockchain where height = (select max(height) from blockchain)").First(&bs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, yerror.ErrBlockNotFound
		}
		return nil, err
	}
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

func (bc *BlockChain) GetRangeBlocks(startHeight, endHeight BlockNum) (blocks []*Block, err error) {
	var bss []BlocksScheme
	err = bc.chain.Db().Where("height BETWEEN ? AND ?", startHeight, endHeight).Find(&bss).Error
	if err != nil {
		return
	}
	compactblocks := bssToBlocks(bss)
	for _, compactblock := range compactblocks {
		var txns SignedTxns
		for _, txnHash := range compactblock.TxnsHashes {
			txn, err := bc.ItxDB.GetTxn(txnHash)
			if err != nil {
				return nil, err
			}
			txns = append(txns, txn)
		}
		block := &Block{
			Header: compactblock.Header,
			Txns:   txns,
		}
		blocks = append(blocks, block)
	}
	return
}
