package types

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types/goproto"
	"unsafe"
)

type SignedTxn struct {
	Raw       *UnsignedTxn
	TxnHash   Hash
	Address   []byte
	Pubkey    []byte
	Signature []byte
	FromP2P   bool
}

type TxnChecker interface {
	CheckTxn(*SignedTxn) error
}

func NewSignedTxn(wrCall *WrCall, pubkey, addr, sig []byte) (*SignedTxn, error) {
	raw, err := NewUnsignedTxn(wrCall)
	if err != nil {
		return nil, err
	}
	stx := &SignedTxn{
		Raw:       raw,
		Address:   addr,
		Pubkey:    pubkey,
		Signature: sig,
	}
	stx.TxnHash, err = stx.GenerateHash()
	if err != nil {
		return nil, err
	}
	return stx, nil
}

func (st *SignedTxn) ChainID() uint64 {
	return st.Raw.WrCall.ChainID
}

func (st *SignedTxn) BindJson(v any) error {
	return BindJsonParams(st.Raw.WrCall.Params, v)
}

func (st *SignedTxn) GetParams() string {
	return st.Raw.WrCall.Params
}

func (st *SignedTxn) GetTips() uint64 {
	return st.Raw.WrCall.Tips
}

func (st *SignedTxn) GetLeiPrice() uint64 {
	return st.Raw.WrCall.LeiPrice
}

func (st *SignedTxn) ParamsIsJson() bool {
	return json.Valid([]byte(st.GetParams()))
}

func (st *SignedTxn) SetParams(params string) {
	st.Raw.WrCall.Params = params
}

func (st *SignedTxn) GetCaller() *Address {
	addr := BytesToAddress(st.Address)
	return &addr
}

func (st *SignedTxn) GetEthFormatCaller() *Address {
	if st.Pubkey == nil {
		return nil
	}
	var addr Address
	addrByt := BytesToAddress(Keccak256(st.Pubkey[1:])[12:])
	copy(addr[:], addrByt.Bytes())
	return &addr
}

func (st *SignedTxn) TripodName() string {
	return st.Raw.WrCall.TripodName
}

func (st *SignedTxn) WrName() string {
	return st.Raw.WrCall.FuncName
}

func (st *SignedTxn) BindJsonParams(v interface{}) error {
	return st.Raw.BindJsonParams(v)
}

func (st *SignedTxn) FromP2p() bool {
	return st.FromP2P
}

func (st *SignedTxn) ToPb() *goproto.SignedTxn {
	return &goproto.SignedTxn{
		Raw:       st.Raw.ToPb(),
		TxnHash:   st.TxnHash.Bytes(),
		Pubkey:    st.Pubkey,
		Signature: st.Signature,
	}
}

func SignedTxnFromPb(pb *goproto.SignedTxn) (*SignedTxn, error) {
	return &SignedTxn{
		Raw:       UnsignedTxnFromPb(pb.Raw),
		TxnHash:   BytesToHash(pb.TxnHash),
		Pubkey:    pb.Pubkey,
		Signature: pb.Signature,
	}, nil
}

func (st *SignedTxn) GenerateHash() (Hash, error) {
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
	WrCall    *WrCall
	Timestamp uint64
	// Nonce is unnecessary
	Nonce uint64
}

func NewUnsignedTxn(wrCall *WrCall) (*UnsignedTxn, error) {
	return &UnsignedTxn{
		WrCall: wrCall,
		// Timestamp: ytime.NowNanoTsU64(),
	}, nil
}

func (ut *UnsignedTxn) BindJsonParams(v interface{}) error {
	return ut.WrCall.BindJsonParams(v)
}

func (ut *UnsignedTxn) ToPb() *goproto.UnsignedTxn {
	return &goproto.UnsignedTxn{
		WrCall: &goproto.WrCall{
			ChainId:    ut.WrCall.ChainID,
			TripodName: ut.WrCall.TripodName,
			FuncName:   ut.WrCall.FuncName,
			Params:     ut.WrCall.Params,
			LeiPrice:   ut.WrCall.LeiPrice,
			Tips:       ut.WrCall.Tips,
		},
		Timestamp: ut.Timestamp,
	}
}

func UnsignedTxnFromPb(pb *goproto.UnsignedTxn) *UnsignedTxn {
	return &UnsignedTxn{
		WrCall: &WrCall{
			ChainID:    pb.WrCall.ChainId,
			TripodName: pb.WrCall.TripodName,
			FuncName:   pb.WrCall.FuncName,
			Params:     pb.WrCall.Params,
			LeiPrice:   pb.WrCall.LeiPrice,
			Tips:       pb.WrCall.Tips,
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
