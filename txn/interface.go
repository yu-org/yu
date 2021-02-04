package txn

import (
	. "yu/common"
	"yu/event"
	. "yu/keypair"
)

type Itxn interface {
	Events() []event.IEvent
	Caller() PubKey
	Ecalls() []*Ecall
	Timestamp() int64
	Hash() (Hash, error)
	Sign(key PrivKey) error
	Verify() error
	Encode() ([]byte, error)
	Extra() interface{}
}
