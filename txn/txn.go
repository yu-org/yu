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
	caller Address
	calls  []*Call
	events []event.Event
	signature []byte
}

func NewTxn(caller Address, calls []*Call) (*Txn, error) {
	txn := &Txn{
		caller: caller,
		calls:  calls,
		events: make([]event.Event, 0),
		signature: nil,
	}
	id, err := txn.Hash()
	if err != nil {
		return nil, err
	}
	txn.id = id
	return txn, nil
}

func (t *Txn) Events() []event.Event {
	return t.events
}

func (t *Txn) Caller() Address {
	return t.caller
}

func (t *Txn) Calls() []*Call {
	return t.calls
}

func (t *Txn) Hash() (Hash, error) {
	var hash Hash
	byt, err := t.Encode()
	if err != nil {
		return NullHash, err
	}
	hash = sha256.Sum256(byt)
	return hash, nil
}

func (t *Txn) Sign(key KeyPair) (err error) {
	// Notice:  Use Encoder of the txn or Hash?
	var data Hash
	data, err = t.Hash()
	if err != nil {
		return
	}
	t.signature, err = key.SignData(data.Bytes())
	return
}

func (t *Txn) Verify(key KeyPair) (bool, error) {
	// Notice:  Use Encoder of the txn or Hash?
	data, err := t.Hash()
	if err != nil {
		return false, err
	}
	return key.VerifySigner(data.Bytes(), t.signature), nil
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
	return t.signature != nil
}
