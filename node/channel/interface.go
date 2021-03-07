package channel

import (
	. "yu/blockchain"
	. "yu/node"
	. "yu/txn"
)

type BlockChannel interface {
	Channel() chan IBlock
	Push(IBlock) error
	Pop() (IBlock, error)
}

type TxnsChannel interface {
	Push(IsignedTxn) error
	Pop(num int) (SignedTxns, error)
}

type TransferBodyChannel interface {
	Push(*TransferBody) error
	Pop() (*TransferBody, error)
}
