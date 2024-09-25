package blockchain

import (
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	ysql "github.com/yu-org/yu/infra/storage/sql"
	"sync/atomic"
)

type BlockChain struct {
	nodeType int

	chainID uint64

	currentBlock       atomic.Pointer[Block]
	lastFinalizedBlock atomic.Pointer[Block]
	finalizedBlocks    *lru.Cache[BlockNum, *Block]

	chain ysql.SqlDB
	ItxDB
}

func NewBlockChain(nodeType int, cfg *config.BlockchainConf, txdb ItxDB) *BlockChain {
	chain, err := ysql.NewSqlDB(&cfg.ChainDB)
	if err != nil {
		logrus.Fatal("init blockchain SQL db failed: ", err)
	}

	err = chain.AutoMigrate(&BlocksScheme{})
	if err != nil {
		logrus.Fatal("create blockchain scheme failed: ", err)
	}

	var currentBlock, lastFinalizedBlock atomic.Pointer[Block]

	currentBlock.Store(nil)
	lastFinalizedBlock.Store(nil)

	finalizedBlocks, err := lru.New[BlockNum, *Block](cfg.CacheSize)
	if err != nil {
		logrus.Fatal("init cache failed: ", err)
	}

	return &BlockChain{
		nodeType:           nodeType,
		chainID:            cfg.ChainID,
		currentBlock:       currentBlock,
		lastFinalizedBlock: lastFinalizedBlock,
		finalizedBlocks:    finalizedBlocks,
		chain:              chain,
		ItxDB:              txdb,
	}
}

func (bc *BlockChain) ConvergeType() ConvergeType {
	return Longest
}

func (bc *BlockChain) ChainID() uint64 {
	return bc.chainID
}

func (bc *BlockChain) NewEmptyBlock() *Block {
	return &Block{Header: &Header{ChainID: bc.chainID}, Txns: nil}
}

func (bc *BlockChain) GetGenesis() (*Block, error) {
	var block BlocksScheme
	result := bc.chain.Db().Where("height = ?", 0).Find(&block)
	err := result.Error

	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, yerror.ErrBlockNotFound
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
		return bc.appendBlock(b)
	}
	return nil
}

func (bc *BlockChain) AppendBlock(b *Block) error {
	err := bc.appendBlock(b)
	if err != nil {
		return err
	}
	bc.currentBlock.Store(b)
	return nil
}

func (bc *BlockChain) appendBlock(b *Block) error {
	//start := time.Now()
	//defer func() {
	//	metrics.AppendBlockDuration.WithLabelValues(strconv.FormatInt(int64(b.Height), 10)).Observe(time.Since(start).Seconds())
	//}()
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
	result := bc.chain.Db().Where(&BlocksScheme{
		Hash: blockHash.String(),
	}).Find(&bs)
	err := result.Error

	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, yerror.ErrBlockNotFound
	}
	return bs.toBlock()
}

func (bc *BlockChain) GetBlock(blockHash Hash) (*Block, error) {
	cBlock, err := bc.GetCompactBlock(blockHash)
	if err != nil {
		return nil, err
	}
	return bc.getBlockByCompact(cBlock)
}

func (bc *BlockChain) GetCompactBlockByHeight(height BlockNum) (*CompactBlock, error) {
	if block, ok := bc.finalizedBlocks.Get(height); ok {
		return block.Compact(), nil
	}
	var bs BlocksScheme
	result := bc.chain.Db().Where(&BlocksScheme{
		Height:   height,
		Finalize: true,
	}).Find(&bs)
	err := result.Error

	if err != nil {
		return nil, err
	}

	if result.RowsAffected == 0 {
		return nil, yerror.ErrBlockNotFound
	}
	return bs.toBlock()
}

func (bc *BlockChain) GetBlockByHeight(height BlockNum) (*Block, error) {
	if block, ok := bc.finalizedBlocks.Get(height); ok {
		return block, nil
	}
	cBlock, err := bc.GetCompactBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	return bc.getBlockByCompact(cBlock)
}

func (bc *BlockChain) GetAllBlocksByHeight(height BlockNum) ([]*Block, error) {
	cBlocks, err := bc.GetAllCompactBlocksByHeight(height)
	if err != nil {
		return nil, err
	}
	var blocks []*Block
	for _, cBlock := range cBlocks {
		block, err := bc.getBlockByCompact(cBlock)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (bc *BlockChain) GetAllCompactBlocksByHeight(height BlockNum) ([]*CompactBlock, error) {
	var bss []BlocksScheme
	err := bc.chain.Db().Where(&BlocksScheme{Height: height}).Find(&bss).Error
	if err != nil {
		return nil, err
	}
	return bssToBlocks(bss), nil
}

func (bc *BlockChain) UpdateBlock(b *Block) error {
	bs, err := toBlocksScheme(b.Compact())
	if err != nil {
		return err
	}

	err = bc.ItxDB.SetTxns(b.Txns)
	if err != nil {
		return err
	}

	return bc.chain.Db().Where(&BlocksScheme{
		Hash: b.Hash.String(),
	}).Updates(bs).Error
}

func (bc *BlockChain) UpdateBlockByHeight(b *Block) error {
	bs, err := toBlocksScheme(b.Compact())
	if err != nil {
		return err
	}

	err = bc.ItxDB.SetTxns(b.Txns)
	if err != nil {
		return err
	}

	return bc.chain.Db().Where(&BlocksScheme{
		Height: b.Height,
	}).Updates(bs).Error
}

func (bc *BlockChain) Children(prevBlockHash Hash) ([]*Block, error) {
	cBlocks, err := bc.ChildrenCompact(prevBlockHash)
	if err != nil {
		return nil, err
	}
	var blocks []*Block
	for _, cBlock := range cBlocks {
		block, err := bc.getBlockByCompact(cBlock)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (bc *BlockChain) ChildrenCompact(prevBlockHash Hash) ([]*CompactBlock, error) {
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

func (bc *BlockChain) Finalize(block *Block) error {
	err := bc.chain.Db().Model(&BlocksScheme{}).Where(&BlocksScheme{
		Hash: block.Hash.String(),
	}).Updates(BlocksScheme{Finalize: true}).Error
	if err != nil {
		return err
	}
	bc.lastFinalizedBlock.Store(block)
	bc.finalizedBlocks.Add(block.Height, block)
	return nil
}

func (bc *BlockChain) LastFinalizedCompact() (*CompactBlock, error) {
	block := bc.lastFinalizedBlock.Load()
	if block != nil {
		return block.Compact(), nil
	}
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

func (bc *BlockChain) LastFinalized() (*Block, error) {
	block := bc.lastFinalizedBlock.Load()
	if block != nil {
		return block, nil
	}
	cBlock, err := bc.LastFinalizedCompact()
	if err != nil {
		return nil, err
	}
	return bc.getBlockByCompact(cBlock)
}

func (bc *BlockChain) GetEndCompactBlock() (*CompactBlock, error) {
	block := bc.currentBlock.Load()
	if block == nil {
		compactBlock, err := bc.getEndCompactBlockFromDB()
		if err != nil {
			return nil, err
		}
		b, err := bc.getBlockByCompact(compactBlock)
		if err != nil {
			return nil, err
		}
		bc.currentBlock.Store(b)
		return compactBlock, nil
	}
	return block.Compact(), nil
}

func (bc *BlockChain) getEndCompactBlockFromDB() (*CompactBlock, error) {
	var bs BlocksScheme
	result := bc.chain.Db().Raw("select * from blockchain where height = (select max(height) from blockchain)").Find(&bs)
	err := result.Error
	if err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, yerror.ErrBlockNotFound
	}
	return bs.toBlock()
}

func (bc *BlockChain) GetEndBlock() (*Block, error) {
	block := bc.currentBlock.Load()
	if block == nil {
		compactBlock, err := bc.getEndCompactBlockFromDB()
		if err != nil {
			return nil, err
		}
		block, err = bc.getBlockByCompact(compactBlock)
		if err != nil {
			return nil, err
		}
		bc.currentBlock.Store(block)
	}
	return block, nil
}

func (bc *BlockChain) GetAllCompactBlocks() ([]*CompactBlock, error) {
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

func (bc *BlockChain) getBlockByCompact(cBlock *CompactBlock) (*Block, error) {
	txns, err := bc.ItxDB.GetTxns(cBlock.TxnsHashes)
	if err != nil {
		return nil, err
	}
	return &Block{
		Header: cBlock.Header,
		Txns:   txns,
	}, nil
}
