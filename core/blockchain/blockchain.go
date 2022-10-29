package blockchain

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	ysql "github.com/yu-org/yu/infra/storage/sql"
)

type BlockChain struct {
	chain ysql.SqlDB
	txns  ItxDB
}

func NewBlockChain(cfg *config.BlockchainConf, txdb ItxDB) *BlockChain {
	chain, err := ysql.NewSqlDB(&cfg.ChainDB)
	if err != nil {
		logrus.Fatal("init blockchain SQL db error: ", err)
	}

	err = chain.CreateIfNotExist(&BlocksScheme{})
	if err != nil {
		logrus.Fatal("create blockchain scheme: ", err)
	}

	return &BlockChain{
		chain: chain,
		txns:  txdb,
	}
}

func (bc *BlockChain) ConvergeType() ConvergeType {
	return Longest
}

func (bc *BlockChain) NewEmptyBlock() *Block {
	return &Block{Header: &Header{}, Txns: nil}
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
		return bc.appendCompactBlock(b)
	}
	return nil
}

func (bc *BlockChain) AppendBlock(b *Block) error {
	err := bc.appendCompactBlock(b.Compact())
	if err != nil {
		return err
	}
	return bc.txns.SetTxns(b.Txns)
}

func (bc *BlockChain) appendCompactBlock(b *CompactBlock) error {
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
	bc.chain.Db().Model(&BlocksScheme{}).Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Updates(BlocksScheme{Finalize: true})
	return nil
}

func (bc *BlockChain) LastFinalized() (*CompactBlock, error) {
	var bss []BlocksScheme
	bc.chain.Db().Model(&BlocksScheme{}).Where(&BlocksScheme{
		Finalize: true,
	}).Order("height").Find(&bss)
	bs := bss[len(bss)-1]
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

func (bc *BlockChain) GetRangeBlocks(startHeight, endHeight BlockNum) (blocks []*Block, err error) {
	var bss []BlocksScheme
	bc.chain.Db().Where("height BETWEEN ? AND ?", startHeight, endHeight).Find(&bss)
	compactblocks := bssToBlocks(bss)
	for _, compactblock := range compactblocks {
		var txns SignedTxns
		for _, txnHash := range compactblock.TxnsHashes {
			txn, err := bc.txns.GetTxn(txnHash)
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
