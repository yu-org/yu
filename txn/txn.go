package txn

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	. "yu/common"
	"yu/event"
	. "yu/keypair"
)

type Txn struct {
	id     Hash
	caller AccountId
	calls  []*Call
	events []event.Event
}

func NewTxn(caller AccountId, calls []*Call) *Txn {
	return &Txn{
		caller: caller,
		calls:  calls,
		events: make([]event.Event, 0),
	}
}

func (t *Txn) Events() []event.Event {
	return t.events
}

func (t *Txn) Hash() (Hash, error) {
	var hash Hash
	byt, err := t.Encode()
	if err != nil {
		return [HashLen]byte{}, err
	}
	hash = sha256.Sum256(byt)
	return hash, nil
}

func (t *Txn) Sign(key KeyPair) ([]byte, error) {

}

func (t *Txn) Verify(key KeyPair) (bool, error) {

}

func (t *Txn) Check() (bool, error) {

}

func (t *Txn) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(t)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte) (*Txn, error) {
	var txn Txn
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&txn)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (t *Txn) IsSigned() bool {

}
