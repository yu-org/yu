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
	Raw       *UnsignedTxn
	TxnHash   Hash
	Pubkey    PubKey
	Signature []byte
}

func NewSignedTxn(caller Address, ecall *Ecall, pubkey PubKey, sig []byte) (*SignedTxn, error) {
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

func (st *SignedTxn) GetRaw() *UnsignedTxn {
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

func DecodeSignedTxn(data []byte) (st *SignedTxn, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(st)
	return
}

type SignedTxns []*SignedTxn

func FromArray(txns []*SignedTxn) SignedTxns {
	var stxns SignedTxns
	stxns = append(stxns, txns...)
	return stxns
}

func (sts SignedTxns) ToArray() []*SignedTxn {
	return sts[:]
}

func (sts SignedTxns) Encode() ([]byte, error) {
	return GobEncode(sts)
}

func (sts SignedTxns) Decode(data []byte) (SignedTxns, error) {
	err := GobDecode(data, &sts)
	return sts, err
}
