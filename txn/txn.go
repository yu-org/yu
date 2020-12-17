package txn

import (
	"yu/event"
)

type Txn struct {
	events []event.Event
}

func NewTxn() *Txn {
	return &Txn {
		events: make([]event.Event, 0),
	}
}

func(t *Txn) Events() []event.Event {
	return t.events
}

func(t *Txn) Hash() string {

}

func(t *Txn) IsSigned() bool {

}
