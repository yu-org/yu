package txn

type TxPool struct {
	txns []*Txn
}

func NewTxPool() *TxPool {
	return &TxPool{
		txns: make([]*Txn, 0),
	}
}