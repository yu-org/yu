package state

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/config"
)

type StateStore struct {
	KVDB *StateKV
}

func NewStateStore(cfg *StateConf) (*StateStore, error) {
	stateKV, err := NewStateKV(&cfg.KV)
	if err != nil {
		return nil, err
	}
	return &StateStore{KVDB: stateKV}, nil
}

func (ss *StateStore) StartBlock(blockHash Hash) {
	ss.KVDB.StartBlock(blockHash)
}

func (ss *StateStore) SetCanRead(blockHash Hash) {
	ss.KVDB.SetCanRead(blockHash)
}

func (ss *StateStore) Commit() (Hash, error) {
	return ss.KVDB.Commit()
}

func (ss *StateStore) Discard() {
	ss.KVDB.Discard()
}

func (ss *StateStore) DiscardAll() {
	ss.KVDB.DiscardAll()
}

func (ss *StateStore) NextTxn() {
	ss.KVDB.NextTxn()
}
