package blockchain

import (
	"log"
	. "yu/common"
	"yu/storage/kv"
)

// the Key Name of last finalized blockID
var LastFinalizedKey = []byte("Last-Finalized-BlockID")

type BlockChain struct {
	kvdb kv.KV
}

func NewBlockChain(kvType string, kvCfg *kv.KVconf) *BlockChain {
	kvdb, err := kv.NewKV(kvType, kvCfg)
	if err != nil {
		log.Fatal("cannot load kvdb")
	}
	return &BlockChain{
		kvdb: kvdb,
	}
}

func (bc *BlockChain) AppendBlock(ib IBlock) error {
	var b *Block = ib.(*Block)
	blockId := b.BlockId().Bytes()
	blockByt, err := b.Encode()
	if err != nil {
		return err
	}
	return bc.kvdb.Set(blockId, blockByt)
}

func (bc *BlockChain) Children(prevId BlockId) (blocks []IBlock, err error) {
	prevBlockNum, prevHash := prevId.Separate()
	blockNum := prevBlockNum + 1
	iter, err := bc.kvdb.Iter(blockNum.Bytes())
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	for iter.Valid() {
		_, blockByt, err := iter.Entry()
		if err != nil {
			return
		}
		block, err := DecodeBlock(blockByt)
		if err != nil {
			return
		}
		if block.PrevHash() == prevHash {
			blocks = append(blocks, block)
		}

		err = iter.Next()
		if err != nil {
			return
		}
	}
	return
}

func (bc *BlockChain) Finalize(id BlockId) error {
	return bc.kvdb.Set(LastFinalizedKey, id.Bytes())
}

func (bc *BlockChain) LastFinalized() (IBlock, error) {
	lfBlockIdByt, err := bc.kvdb.Get(LastFinalizedKey)
	if err != nil {
		return nil, err
	}
	blockByt, err := bc.kvdb.Get(lfBlockIdByt)
	if err != nil {
		return nil, err
	}
	return DecodeBlock(blockByt)
}

func (bc *BlockChain) Leaves() ([]IBlock, error) {

}
