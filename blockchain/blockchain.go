package blockchain

import (
	"time"
	. "yu/common"
	"yu/config"
	ysql "yu/storage/sql"
	"yu/utils/codec"
	"yu/yerror"
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

func (bc *BlockChain) ConvergeType() ConvergeType {
	return Longest
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

func (bc *BlockChain) EncodeBlocks(blocks []IBlock) ([]byte, error) {
	var bs []*Block
	for _, b := range blocks {
		bs = append(bs, b.(*Block))
	}
	return codec.GobEncode(bs)
}

func (bc *BlockChain) DecodeBlocks(data []byte) ([]IBlock, error) {
	var bs []*Block
	err := codec.GobDecode(data, &bs)
	if err != nil {
		return nil, err
	}
	var blocks []IBlock
	for _, b := range bs {
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func (bc *BlockChain) GetGenesis() (IBlock, error) {
	var block BlocksScheme
	bc.chain.Db().Where("height = ?", 0).First(&block)
	return block.toBlock()
}

func (bc *BlockChain) SetGenesis(b IBlock) error {
	var blocks []BlocksScheme
	bc.chain.Db().Where("height = ?", 0).Find(&blocks)

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

func (bc *BlockChain) TakeP2pBlocksUntil(height BlockNum) (map[BlockNum][]IBlock, error) {
	var bsp []BlocksFromP2pScheme
	bc.blocksFromP2p.Db().Where("height <= ?", height).Order("height").Find(&bsp)
	blocks := bspToBlocks(bsp)
	hBlocks := make(map[BlockNum][]IBlock, 0)
	for _, block := range blocks {
		height := block.GetHeader().GetHeight()
		hBlocks[height] = append(hBlocks[height], block)
	}

	for _, b := range blocks {
		bc.blocksFromP2p.Db().Delete(&BlocksFromP2pScheme{}, b.GetHeader().GetHash())
	}
	return hBlocks, nil
}

func (bc *BlockChain) TakeP2pBlocks(height BlockNum) ([]IBlock, error) {
	var bsp []BlocksFromP2pScheme
	bc.blocksFromP2p.Db().Where("height = ?", height).Find(&bsp)

	for _, bs := range bsp {
		bc.blocksFromP2p.Db().Delete(&BlocksFromP2pScheme{}, bs.BlockHash)
	}
	return bspToBlocks(bsp), nil
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

func (bc *BlockChain) UpdateBlock(b IBlock) error {
	bs, err := toBlocksScheme(b)
	if err != nil {
		return err
	}
	bc.chain.Db().Where(&BlocksScheme{
		Hash: b.GetHeader().GetHash().String(),
	}).Updates(bs)
	return nil
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

func (bc *BlockChain) GetEndBlock() (IBlock, error) {
	chain, err := bc.Chain()
	if err != nil {
		return nil, err
	}

	return chain.Last(), nil
}

func (bc *BlockChain) GetAllBlocks() ([]IBlock, error) {
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

func (bc *BlockChain) GetRangeBlocks(startHeight, endHeight BlockNum) ([]IBlock, error) {
	var bss []BlocksScheme
	bc.chain.Db().Where("height BETWEEN ? AND ?", startHeight, endHeight).Find(&bss)
	return bssToBlocks(bss), nil
}

func (bc *BlockChain) Chain() (IChainStruct, error) {
	switch bc.ConvergeType() {
	case Longest:
		chains, err := bc.LongestChains()
		if err != nil {
			return nil, err
		}
		return chains[0], nil
	case Heaviest:
		chains, err := bc.HeaviestChains()
		if err != nil {
			return nil, err
		}
		return chains[0], nil
	case Finalize:
		return bc.FinalizedChain()
	default:
		return nil, yerror.NoConvergeType
	}
}

func (bc *BlockChain) LongestChains() ([]IChainStruct, error) {
	blocks, err := bc.GetAllBlocks()
	if err != nil {
		return nil, err
	}
	return MakeLongestChain(blocks), nil
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
