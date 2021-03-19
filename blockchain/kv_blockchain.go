package blockchain

import (
	"github.com/sirupsen/logrus"
	"time"
	. "yu/common"
	. "yu/config"
	"yu/storage/kv"
	"yu/storage/queue"
)

// the Key Name of last finalized blockID
var LastFinalizedKey = []byte("Last-Finalized-BlockID")
var PendingBlocksTopic = "pending-blocks"

type BlockChain struct {
	chain         kv.KV
	pendingBlocks queue.Queue
}

func NewKvBlockChain(kvCfg *KVconf, queueCfg *QueueConf) *BlockChain {
	kvdb, err := kv.NewKV(kvCfg)
	if err != nil {
		logrus.Panicln("cannot load chain")
	}
	q, err := queue.NewQueue(queueCfg)
	if err != nil {
		logrus.Panicln("cannot load pending-blocks")
	}
	return &BlockChain{
		chain:         kvdb,
		pendingBlocks: q,
	}
}

func (bc *BlockChain) NewDefaultBlock() IBlock {
	header := &Header{
		timestamp: time.Now().UnixNano(),
	}
	return &Block{
		header: header,
	}
}

func (bc *BlockChain) NewEmptyBlock() IBlock {
	return &Block{}
}

func (bc *BlockChain) PendBlock(ib IBlock) error {
	blockByt, err := ib.Encode()
	if err != nil {
		return err
	}
	return bc.pendingBlocks.Push(PendingBlocksTopic, blockByt)
}

func (bc *BlockChain) PopBlock() (IBlock, error) {
	blockByt, err := bc.pendingBlocks.Pop(PendingBlocksTopic)
	if err != nil {
		return nil, err
	}
	return bc.NewEmptyBlock().Decode(blockByt)
}

func (bc *BlockChain) AppendBlock(ib IBlock) error {
	var b *Block = ib.(*Block)
	blockId := b.BlockId().Bytes()
	blockByt, err := b.Encode()
	if err != nil {
		return err
	}
	return bc.chain.Set(blockId, blockByt)
}

func (bc *BlockChain) GetBlock(id BlockId) (IBlock, error) {
	blockByt, err := bc.chain.Get(id.Bytes())
	if err != nil {
		return nil, err
	}
	return bc.NewEmptyBlock().Decode(blockByt)
}

func (bc *BlockChain) Children(prevId BlockId) ([]IBlock, error) {
	prevBlockNum, prevHash := prevId.Separate()
	blockNum := prevBlockNum + 1
	iter, err := bc.chain.Iter(blockNum.Bytes())
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
		if block.Header().PrevHash() == prevHash {
			blocks = append(blocks, block)
		}

		err = iter.Next()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

func (bc *BlockChain) Finalize(id BlockId) error {
	return bc.chain.Set(LastFinalizedKey, id.Bytes())
}

func (bc *BlockChain) LastFinalized() (IBlock, error) {
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

func (bc *BlockChain) Leaves() ([]IBlock, error) {
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

func (bc *BlockChain) Longest() ([]*ChainStruct, error) {
	blocks, err := bc.Leaves()
	if err != nil {
		return nil, err
	}
	return MakeLongestChain(blocks), nil
}

func (bc *BlockChain) Heaviest() ([]*ChainStruct, error) {
	blocks, err := bc.Leaves()
	if err != nil {
		return nil, err
	}
	return MakeHeaviestChain(blocks), nil
}
