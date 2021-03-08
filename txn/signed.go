package txn

import (
	"bytes"
	"encoding/gob"
	"unsafe"
	. "yu/common"
	. "yu/keypair"
)

type SignedTxn struct {
	Raw       IunsignedTxn
	TxnHash   Hash
	Pubkey    PubKey
	Signature []byte
}

func NewSignedTxn(caller Address, ecall *Ecall, pubkey PubKey, sig []byte) (IsignedTxn, error) {
	raw, err := NewUnsignedTxn(caller, ecall)
	if err != nil {
		return nil, err
	}
	hash, err := raw.Hash()
	if err != nil {
		return nil, err
	}
	return &SignedTxn{
		Raw:       raw,
		TxnHash:   hash,
		Pubkey:    pubkey,
		Signature: sig,
	}, nil
}

func (st *SignedTxn) GetRaw() IunsignedTxn {
	return st.Raw
}

func (st *SignedTxn) GetTxnHash() Hash {
	return st.TxnHash
}

func (st *SignedTxn) GetPubkey() PubKey {
	return st.Pubkey
}

func (st *SignedTxn) GetSignature() []byte {
	return st.Signature
}

func (st *SignedTxn) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(st)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (st *SignedTxn) Size() int {
	return int(unsafe.Sizeof(st))
}

func DecodeSignedTxn(data []byte) (st IsignedTxn, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(st)
	return
}
