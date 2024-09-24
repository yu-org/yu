package state

import "github.com/yu-org/yu/core/types"

type NoStateDB struct {
}

func (n *NoStateDB) Set(triName NameString, key, value []byte) {

}

func (n *NoStateDB) Delete(triName NameString, key []byte) {
}

func (n *NoStateDB) Get(triName NameString, key []byte) ([]byte, error) {
	return nil, nil
}

func (n *NoStateDB) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	return nil, nil
}

func (n *NoStateDB) Exist(triName NameString, key []byte) bool {
	return false
}

func (n *NoStateDB) GetByBlockHash(triName NameString, key []byte, block *types.Block) ([]byte, error) {
	return nil, nil
}

func (n *NoStateDB) Commit() ([]byte, error) {
	return nil, nil
}

func (n *NoStateDB) NextTxn() {
}

func (n *NoStateDB) Discard() {
}

func (n *NoStateDB) DiscardAll() {

}

func (n *NoStateDB) StartBlock(block *types.Block) {
}

func (n *NoStateDB) FinalizeBlock(block *types.Block) {
}
