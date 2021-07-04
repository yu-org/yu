package txn

import (
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/keypair"
	. "github.com/Lawliet-Chan/yu/utils/codec"
	"unsafe"
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
	return GlobalCodec.EncodeToBytes(st)
}

func (st *SignedTxn) Size() int {
	return int(unsafe.Sizeof(st))
}

//func DecodeSignedTxn(data []byte) (st *SignedTxn, err error) {
//	decoder := gob.NewDecoder(bytes.NewReader(data))
//	err = decoder.Decode(st)
//	return
//}

type SignedTxns []*SignedTxn

func FromArray(txns ...*SignedTxn) SignedTxns {
	var stxns SignedTxns
	stxns = append(stxns, txns...)
	return stxns
}

func (sts SignedTxns) ToArray() []*SignedTxn {
	return sts[:]
}

func (sts SignedTxns) Hashes() (hashes []Hash) {
	for _, st := range sts {
		hashes = append(hashes, st.TxnHash)
	}
	return
}

func (sts SignedTxns) Remove(hash Hash) (int, SignedTxns) {
	for i, stxn := range sts {
		if stxn.GetTxnHash() == hash {
			if i == 0 {
				sts = sts[1:]
				return i, sts
			}
			if i == len(sts)-1 {
				sts = sts[:i]
				return i, sts
			}

			sts = append(sts[:i], sts[i+1:]...)
			return i, sts
		}
	}
	return -1, nil
}

func (sts SignedTxns) Encode() ([]byte, error) {
	var msts MidSignedTxns
	for _, st := range sts {
		msts = append(msts, &MidSignedTxn{
			Raw:       st.Raw,
			TxnHash:   st.TxnHash,
			KeyType:   st.Pubkey.Type(),
			Pubkey:    st.Pubkey.Bytes(),
			Signature: st.Signature,
		})
	}
	return GlobalCodec.EncodeToBytes(msts)
}

func DecodeSignedTxns(data []byte) (SignedTxns, error) {
	mtxns := MidSignedTxns{}
	err := GlobalCodec.DecodeBytes(data, &mtxns)
	if err != nil {
		return nil, err
	}
	var sts SignedTxns
	for _, mtxn := range mtxns {
		pubKey, err := PubKeyFromBytes(mtxn.KeyType, mtxn.Pubkey)
		if err != nil {
			return nil, err
		}
		sts = append(sts, &SignedTxn{
			Raw:       mtxn.Raw,
			TxnHash:   mtxn.TxnHash,
			Pubkey:    pubKey,
			Signature: mtxn.Signature,
		})
	}
	return sts, err
}

type MidSignedTxns []*MidSignedTxn

type MidSignedTxn struct {
	Raw       *UnsignedTxn
	TxnHash   Hash
	KeyType   string
	Pubkey    []byte
	Signature []byte
}
