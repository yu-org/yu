package txn

type TxPool struct {
	poolSize uint64
	txns []*Txn
}

func NewTxPool(poolSize uint64) *TxPool {
	return &TxPool{
		poolSize: poolSize,
		txns: make([]*Txn, 0),
	}
}

func(tp *TxPool) PutTxn(txn *Txn) bool {
	tp.txns = append(tp.txns, txn)
}

func (tp *TxPool) checkTxn(txn *Txn) bool {

}