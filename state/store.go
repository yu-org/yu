package state

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/config"
)

type StateStore struct {
	KVDB *StateKV
}

func NewStateStore(cfg *StateConf, canRead Hash) (*StateStore, error) {
	stateKV, err := NewStateKV(&cfg.KvDB, canRead)
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

func (ss *StateStore) Commit() error {
	return ss.KVDB.Commit()
}

func (ss *StateStore) Discard() {
	ss.KVDB.Discard()
}
