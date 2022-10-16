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

type TxnChecker interface {
	CheckTxn(*SignedTxn) error
}

func NewSignedTxn(caller Address, ecall *WrCall, pubkey PubKey, sig []byte) (*SignedTxn, error) {
	raw, err := NewUnsignedTxn(caller, ecall)
	if err != nil {
		return nil, err
	}
	stx := &SignedTxn{
		Raw:       raw,
		Pubkey:    pubkey,
		Signature: sig,
	}
	stx.TxnHash, err = stx.Hash()
	if err != nil {
		return nil, err
	}
	return stx, nil
}

func (st *SignedTxn) BindJsonParams(v interface{}) error {
	return st.Raw.BindJsonParams(v)
}

func (st *SignedTxn) ToPb() *goproto.SignedTxn {
	return &goproto.SignedTxn{
		Raw:       st.Raw.ToPb(),
		TxnHash:   st.TxnHash.Bytes(),
		Pubkey:    st.Pubkey.BytesWithType(),
		Signature: st.Signature,
	}
}

func SignedTxnFromPb(pb *goproto.SignedTxn) (*SignedTxn, error) {
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

func (st *SignedTxn) Hash() (Hash, error) {
	var hash Hash
	byt, err := st.Encode()
	if err != nil {
		return NullHash, err
	}
	hash = sha256.Sum256(byt)
	return hash, nil
}

func (st *SignedTxn) Encode() ([]byte, error) {
	return proto.Marshal(st.ToPb())
}

func (st *SignedTxn) Size() int {
	return int(unsafe.Sizeof(st))
}

func DecodeSignedTxn(data []byte) (st *SignedTxn, err error) {
	var pb goproto.SignedTxn
	err = proto.Unmarshal(data, &pb)
	if err != nil {
		return nil, err
	}
	return SignedTxnFromPb(&pb)
}

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
		txn, err := SignedTxnFromPb(tx)
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
	Caller    Address
	WrCall    *WrCall
	Timestamp uint64
	// Nonce is unnecessary
	Nonce uint64
}

func NewUnsignedTxn(caller Address, wrCall *WrCall) (*UnsignedTxn, error) {
	return &UnsignedTxn{
		Caller:    caller,
		WrCall:    wrCall,
		Timestamp: ytime.NowNanoTsU64(),
	}, nil
}

func (ut *UnsignedTxn) BindJsonParams(v interface{}) error {
	return ut.WrCall.BindJsonParams(v)
}

func (ut *UnsignedTxn) ToPb() *goproto.UnsignedTxn {
	return &goproto.UnsignedTxn{
		Caller: ut.Caller.Bytes(),
		Ecall: &goproto.Ecall{
			TripodName: ut.WrCall.TripodName,
			ExecName:   ut.WrCall.ExecName,
			Params:     ut.WrCall.Params,
			LeiPrice:   ut.WrCall.LeiPrice,
		},
		Timestamp: ut.Timestamp,
	}
}

func UnsignedTxnFromPb(pb *goproto.UnsignedTxn) *UnsignedTxn {
	return &UnsignedTxn{
		Caller: BytesToAddress(pb.Caller),
		WrCall: &WrCall{
			TripodName: pb.Ecall.TripodName,
			ExecName:   pb.Ecall.ExecName,
			Params:     pb.Ecall.Params,
			LeiPrice:   pb.Ecall.LeiPrice,
		},
		Timestamp: pb.Timestamp,
	}
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
