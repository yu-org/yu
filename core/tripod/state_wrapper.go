package tripod

import (
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
)

func (t *Tripod) Set(key, value []byte) {
	t.State.Set(t, key, value)
}

func (t *Tripod) Get(key []byte) ([]byte, error) {
	return t.State.Get(t, key)
}

func (t *Tripod) Delete(key []byte) {
	t.State.Delete(t, key)
}

func (t *Tripod) GetFinalized(key []byte) ([]byte, error) {
	return t.State.GetFinalized(t, key)
}

func (t *Tripod) Exist(key []byte) bool {
	return t.State.Exist(t, key)
}

func (t *Tripod) GetByBlockHash(key []byte, block *types.Block) ([]byte, error) {
	return t.State.GetByBlockHash(t, key, block)
}

func (t *Tripod) NextTxn() {
	t.State.NextTxn()
}

func (t *Tripod) CommitState() (common.Hash, error) {
	hash, err := t.State.Commit()
	if err != nil {
		return common.NullHash, err
	}
	return common.BytesToHash(hash), nil
}

func (t *Tripod) Discard() {
	t.State.Discard()
}

func (t *Tripod) DiscardAll() {
	t.State.DiscardAll()
}
