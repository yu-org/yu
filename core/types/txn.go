package types

import (
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/types/goproto"
	ytime "github.com/yu-org/yu/utils/time"
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

func (st *SignedTxn) ToPb() *goproto.SignedTxn {
	return &goproto.SignedTxn{
		Raw:       st.Raw.ToPb(),
		TxnHash:   st.TxnHash.Bytes(),
		Pubkey:    st.Pubkey.BytesWithType(),
		Signature: st.Signature,
	}
}

func SignedTxFromPb(pb *goproto.SignedTxn) (*SignedTxn, error) {
	pubkey, err := PubKeyFromBytes(pb.Pubkey)
	if err != nil {
		return nil, err
	}
	return &SignedTxn{
		Raw:       UnsignedTxnFromPb(pb.Raw),
		TxnHash:   BytesToHash(pb.TxnHash),
		Pubkey:    pubkey,
		Signature: pb.Signature,
	}, nil
}

func (st *SignedTxn) Encode() ([]byte, error) {
	return proto.Marshal(st.ToPb())
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

func (sts SignedTxns) ToPb() *goproto.SignedTxns {
	var pbTxns []*goproto.SignedTxn
	for _, st := range sts {
		pbTxns = append(pbTxns, st.ToPb())
	}
	return &goproto.SignedTxns{Txns: pbTxns}
}

func SignedTxnsFromPb(pb *goproto.SignedTxns) (SignedTxns, error) {
	var sts SignedTxns
	for _, tx := range pb.Txns {
		txn, err := SignedTxFromPb(tx)
		if err != nil {
			return nil, err
		}
		sts = append(sts, txn)
	}
	return sts, nil
}

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
		if stxn.TxnHash == hash {
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
	return proto.Marshal(sts.ToPb())
}

func DecodeSignedTxns(data []byte) (SignedTxns, error) {
	var pb goproto.SignedTxns
	err := proto.Unmarshal(data, &pb)
	if err != nil {
		return nil, err
	}
	return SignedTxnsFromPb(&pb)
}

type UnsignedTxn struct {
	Id        Hash
	Caller    Address
	Ecall     *Ecall
	LeiPrice  uint64
	Timestamp uint64
}

func NewUnsignedTxn(caller Address, ecall *Ecall) (*UnsignedTxn, error) {
	utxn := &UnsignedTxn{
		Caller:    caller,
		Ecall:     ecall,
		Timestamp: ytime.NowNanoTsU64(),
	}
	id, err := utxn.Hash()
	if err != nil {
		return nil, err
	}
	utxn.Id = id
	return utxn, nil
}

func (ut *UnsignedTxn) ToPb() *goproto.UnsignedTxn {
	return &goproto.UnsignedTxn{
		Id:       ut.Id.Bytes(),
		Caller:   ut.Caller.Bytes(),
		LeiPrice: ut.LeiPrice,
		Ecall: &goproto.Ecall{
			TripodName: ut.Ecall.TripodName,
			ExecName:   ut.Ecall.ExecName,
			Params:     ut.Ecall.Params,
		},
		Timestamp: ut.Timestamp,
	}
}

func UnsignedTxnFromPb(pb *goproto.UnsignedTxn) *UnsignedTxn {
	return &UnsignedTxn{
		Id:       BytesToHash(pb.Id),
		Caller:   BytesToAddress(pb.Caller),
		LeiPrice: pb.LeiPrice,
		Ecall: &Ecall{
			TripodName: pb.Ecall.TripodName,
			ExecName:   pb.Ecall.ExecName,
			Params:     pb.Ecall.Params,
		},
		Timestamp: 0,
	}
}

func (ut *UnsignedTxn) Hash() (Hash, error) {
	var hash Hash
	byt, err := ut.Encode()
	if err != nil {
		return NullHash, err
	}
	hash = sha256.Sum256(byt)
	return hash, nil
}

func (ut *UnsignedTxn) Encode() ([]byte, error) {
	return proto.Marshal(ut.ToPb())
}

func DecodeUnsignedTxn(data []byte) (*UnsignedTxn, error) {
	var pb goproto.UnsignedTxn
	err := proto.Unmarshal(data, &pb)
	if err != nil {
		return nil, err
	}
	return UnsignedTxnFromPb(&pb), nil
}
