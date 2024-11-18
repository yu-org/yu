package state

import (
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/storage/kv"
)

type SampleKV struct {
	kvdb kv.KV
}

func (s *SampleKV) Set(triName NameString, key, value []byte) {
	//TODO implement me
	tx, _ := s.kvdb.NewKvTxn()

	panic("implement me")
}

func (s *SampleKV) Delete(triName NameString, key []byte) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) Get(triName NameString, key []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) GetFinalized(triName NameString, key []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) Exist(triName NameString, key []byte) bool {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) GetByBlockHash(triName NameString, key []byte, block *types.Block) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) Commit() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) NextTxn() {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) Discard() {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) DiscardAll() {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) StartBlock(block *types.Block) {
	//TODO implement me
	panic("implement me")
}

func (s *SampleKV) FinalizeBlock(block *types.Block) {
	//TODO implement me
	panic("implement me")
}
