package txn

import (
	"yu/event"
	"yu/txn/context"
)

type TxnFn func(ctx *context.Context) error

type Txn struct {
	txnFns []TxnFn
	events []event.Event
}

func NewTxn() *Txn {
	return &Txn {
		txnFns: make([]TxnFn, 0),
		events: make([]event.Event, 0),
	}
}

func(t *Txn) SetTxnFns(fns ...TxnFn) {
	t.txnFns = append(t.txnFns, fns...)
}

func(t *Txn) Events() []event.Event {
	return t.events
}

func(t *Txn) Hash() string {

}

func(t *Txn) IsSigned() bool {

}
