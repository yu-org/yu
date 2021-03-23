package txn

import (
	"bytes"
	"encoding/gob"
	"unsafe"
	. "yu/common"
	. "yu/keypair"
	. "yu/utils/codec"
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
	return GobEncode(st)
}

func (st *SignedTxn) Size() int {
	return int(unsafe.Sizeof(st))
}

func DecodeSignedTxn(data []byte) (st IsignedTxn, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(st)
	return
}

type SignedTxns []IsignedTxn

func FromArray(txns []IsignedTxn) SignedTxns {
	var stxns SignedTxns
	stxns = append(stxns, txns...)
	return stxns
}

func (sts SignedTxns) ToArray() []IsignedTxn {
	return sts[:]
}

func (sts SignedTxns) Encode() ([]byte, error) {
	return GobEncode(sts)
}

func DecodeSignedTxns(data []byte) (SignedTxns, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var sts SignedTxns
	err := decoder.Decode(&sts)
	if err != nil {
		return nil, err
	}
	return sts, nil
}
