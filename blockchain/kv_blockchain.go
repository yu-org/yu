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
var blocksFromP2pTopic = "blocks-from-p2p"

type KvBlockChain struct {
	chain         kv.KV
	blocksFromP2p queue.Queue
}

func NewKvBlockChain(kvCfg *KVconf, queueCfg *QueueConf) *KvBlockChain {
	kvdb, err := kv.NewKV(kvCfg)
	if err != nil {
		logrus.Panicf("load chain error: %s", err.Error())
	}
	q, err := queue.NewQueue(queueCfg)
	if err != nil {
		logrus.Panicf("load blocks-from-p2p error: %s", err.Error())
	}
	return &KvBlockChain{
		chain:         kvdb,
		blocksFromP2p: q,
	}
}

func (bc *KvBlockChain) NewDefaultBlock() IBlock {
	header := &Header{
		timestamp: time.Now().UnixNano(),
	}
	return &Block{
		header: header,
	}
}

func (bc *KvBlockChain) NewEmptyBlock() IBlock {
	return &Block{}
}

// pending a block from other KvBlockChain-node for validating
func (bc *KvBlockChain) InsertBlockFromP2P(ib IBlock) error {
	blockByt, err := ib.Encode()
	if err != nil {
		return err
	}
	return bc.blocksFromP2p.Push(blocksFromP2pTopic, blockByt)
}

func (bc *KvBlockChain) GetBlockFromP2P() (IBlock, error) {
	blockByt, err := bc.blocksFromP2p.Pop(blocksFromP2pTopic)
	if err != nil {
		return nil, err
	}
	return bc.NewEmptyBlock().Decode(blockByt)
}

func (bc *KvBlockChain) RemoveBlockFromP2P() error {

}

func (bc *KvBlockChain) AppendBlock(b IBlock) error {
	blockId := b.Header().Hash().Bytes()
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
	height := prevBlock.Header().Height() + 1
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
		if block.Header().PrevHash() == prevBlockHash {
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
