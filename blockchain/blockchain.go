package blockchain

import (
	"log"
	. "yu/common"
	"yu/storage/kv"
)

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
	prevBlockNum, _ := prevId.Separate()
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
		block, err := Decode(blockByt)
		if err != nil {
			return
		}
		blocks = append(blocks, block)
		err = iter.Next()
		if err != nil {
			return
		}
	}
	return
}

func (bc *BlockChain) LastFinalized() (IBlock, error) {

}

func (bc *BlockChain) Leaves() ([]IBlock, error) {

}
