package txn

import (
	. "yu/common"
	. "yu/keypair"
)

type SignedTxn struct {
	Raw       *UnsignedTxn
	TxnHash   Hash
	Pubkey    PubKey
	Signature []byte
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
