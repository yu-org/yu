package tripod

import . "github.com/yu-org/yu/common"

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

func (t *Tripod) GetByBlockHash(key []byte, blockHash Hash) ([]byte, error) {
	return t.State.GetByBlockHash(t, key, blockHash)
}

func (t *Tripod) NextTxn() {
	t.State.NextTxn()
}

func (t *Tripod) Commit() (Hash, error) {
	return t.State.Commit()
}

func (t *Tripod) Discard() {
	t.State.Discard()
}

func (t *Tripod) DiscardAll() {
	t.State.DiscardAll()
}

func (t *Tripod) StartBlock(blockHash Hash) {
	t.State.StartBlock(blockHash)
}

func (t *Tripod) FinalizeBlock(blockHash Hash) {
	t.State.FinalizeBlock(blockHash)
}
