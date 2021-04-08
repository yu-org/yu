package blockchain

import (
	"github.com/sirupsen/logrus"
	"time"
	. "yu/common"
	"yu/config"
	"yu/storage/kv"
	ysql "yu/storage/sql"
)

// the Key Name of last finalized blockID
var LastFinalizedKey = []byte("Last-Finalized-BlockID")

type KvBlockChain struct {
	chain         kv.KV
	blocksFromP2p ysql.SqlDB
}

func NewKvBlockChain(cfg *config.BlockchainConf) *KvBlockChain {
	chainKV, err := kv.NewKV(&cfg.ChainKV)
	if err != nil {
		logrus.Panicf("load blockchain KVDB error: %s", err.Error())
	}
	blocksFromP2pDB, err := ysql.NewSqlDB(&cfg.BlocksFromP2pDB)
	if err != nil {
		logrus.Panicf("load blocks-from-p2p SQL-DB error: %s", err.Error())
	}

	return &KvBlockChain{
		chain:         chainKV,
		blocksFromP2p: blocksFromP2pDB,
	}
}

func (bc *KvBlockChain) NewDefaultBlock() IBlock {
	header := &Header{
		Timestamp: time.Now().UnixNano(),
	}
	return &Block{
		Header: header,
	}
}

func (bc *KvBlockChain) NewEmptyBlock() IBlock {
	return &Block{}
}

func (bc *KvBlockChain) SetGenesis(b IBlock) error {
	iter, err := bc.chain.Iter(nil)
	if err != nil {
		return err
	}
	if iter.Valid() {
		return nil
	}
	return bc.AppendBlock(b)
}

// pending a block from other KvBlockChain-node for validating
func (bc *KvBlockChain) InsertBlockFromP2P(b IBlock) error {
	bs, err := toBlocksFromP2pScheme(b)
	if err != nil {
		return err
	}
	bc.blocksFromP2p.Db().Create(&bs)
	return nil
}

func (bc *KvBlockChain) GetBlocksFromP2P(height BlockNum) ([]IBlock, error) {
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

func (bc *KvBlockChain) FlushBlocksFromP2P(height BlockNum) error {
	bc.blocksFromP2p.Db().Where(&BlocksFromP2pScheme{
		Height: height,
	}).Delete(BlocksFromP2pScheme{})
	return nil
}

func (bc *KvBlockChain) AppendBlock(b IBlock) error {
	blockId := b.GetHeader().GetHash().Bytes()
	blockByt, err := b.Encode()
	if err != nil {
		return err
	}
	return bc.chain.Set(blockId, blockByt)
}

func (bc *KvBlockChain) GetBlock(blockHash Hash) (IBlock, error) {
	blockByt, err := bc.chain.Get(blockHash.Bytes())
	if err != nil {
		return nil, err
	}
	return bc.NewEmptyBlock().Decode(blockByt)
}

func (bc *KvBlockChain) Children(prevBlockHash Hash) ([]IBlock, error) {
	prevBlock, err := bc.GetBlock(prevBlockHash)
	if err != nil {
		return nil, err
	}
	height := prevBlock.GetHeader().GetHeight() + 1
	// FIXME: not height for prefix
	iter, err := bc.chain.Iter(height.Bytes())
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var blocks []IBlock
	for iter.Valid() {
		_, blockByt, err := iter.Entry()
		if err != nil {
			return nil, err
		}
		block, err := bc.NewEmptyBlock().Decode(blockByt)
		if err != nil {
			return nil, err
		}
		if block.GetHeader().GetPrevHash() == prevBlockHash {
			blocks = append(blocks, block)
		}

		err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

func (bc *KvBlockChain) Finalize(blockHash Hash) error {
	return bc.chain.Set(LastFinalizedKey, blockHash.Bytes())
}

func (bc *KvBlockChain) LastFinalized() (IBlock, error) {
	lfBlockIdByt, err := bc.chain.Get(LastFinalizedKey)
	if err != nil {
		return nil, err
	}
	blockByt, err := bc.chain.Get(lfBlockIdByt)
	if err != nil {
		return nil, err
	}
	return bc.NewEmptyBlock().Decode(blockByt)
}

func (bc *KvBlockChain) Leaves() ([]IBlock, error) {
	iter, err := bc.chain.Iter(BlockNum(0).Bytes())
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var blocks []IBlock
	for iter.Valid() {
		_, blockByt, err := iter.Entry()
		if err != nil {
			return nil, err
		}
		block, err := bc.NewEmptyBlock().Decode(blockByt)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (bc *KvBlockChain) Longest() ([]IChainStruct, error) {
	blocks, err := bc.Leaves()
	if err != nil {
		return nil, err
	}
	return MakeLongestChain(blocks), nil
}

func (bc *KvBlockChain) Heaviest() ([]IChainStruct, error) {
	blocks, err := bc.Leaves()
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
