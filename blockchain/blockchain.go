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
	blockHash := b.Hash().Bytes()
	blockByt, err := b.Encode()
	if err != nil {
		return err
	}
	return bc.kvdb.Set(blockHash, blockByt)
}

func (bc *BlockChain) Children(hash Hash) []IBlock {

}

func (bc *BlockChain) LastFinalized() IBlock {

}

func (bc *BlockChain) Leaves() []IBlock {

}
