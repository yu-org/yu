package txn

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
	. "yu/common"
	"yu/event"
	. "yu/keypair"
)

type Txn struct {
	id        Hash
	caller    PubKey
	ecalls    []*Ecall
	events    []event.IEvent
	signature []byte
	timestamp int64
	extra     interface{}
}

func NewTxn(caller PubKey, ecalls []*Ecall, extra interface{}) (*Txn, error) {
	txn := &Txn{
		caller:    caller,
		ecalls:    ecalls,
		events:    make([]event.IEvent, 0),
		signature: nil,
		timestamp: time.Now().UnixNano(),
		extra:     extra,
	}
	id, err := txn.Hash()
	if err != nil {
		return nil, err
	}
	txn.id = id
	return txn, nil
}

func (t *Txn) Events() []event.IEvent {
	return t.events
}

func (t *Txn) Caller() PubKey {
	return t.caller
}

func (t *Txn) Ecalls() []*Ecall {
	return t.ecalls
}

func (t *Txn) Timestamp() int64 {
	return t.timestamp
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

func (t *Txn) Sign(key PrivKey) (err error) {
	// Notice:  Use Encoder of the txn or Hash?
	var data Hash
	data, err = t.Hash()
	if err != nil {
		return
	}
	t.signature, err = key.SignData(data.Bytes())
	return
}

func (t *Txn) Verify() error {
	// Notice:  Use Encoder of the txn or Hash?
	data, err := t.Hash()
	if err != nil {
		return err
	}
	if t.caller.VerifySignature(data.Bytes(), t.signature) {
		return nil
	}
	return TxnSignatureErr
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

func (t *Txn) Extra() interface{} {
	return t.extra
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
