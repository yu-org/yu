package txdb

//
//import (
//	. "github.com/yu-org/yu/common"
//	"github.com/yu-org/yu/core/keypair"
//	. "github.com/yu-org/yu/core/receipt"
//	. "github.com/yu-org/yu/core/types"
//	"gorm.io/gorm"
//)
//
//type TxnScheme struct {
//	TxnHash   string `gorm:"primaryKey"`
//	Pubkey    string
//	Signature string
//	RawTxn    string
//	BlockHash string
//}
//
//func (TxnScheme) TableName() string {
//	return "txns"
//}
//
//func newTxnScheme(blockHash Hash, stxn *SignedTxn) (TxnScheme, error) {
//	txnSm, err := toTxnScheme(stxn)
//	if err != nil {
//		return TxnScheme{}, err
//	}
//	txnSm.BlockHash = blockHash.String()
//	return txnSm, nil
//}
//
//func toTxnScheme(stxn *SignedTxn) (TxnScheme, error) {
//	rawTxnByt, err := stxn.Raw.Encode()
//	if err != nil {
//		return TxnScheme{}, err
//	}
//	return TxnScheme{
//		TxnHash:   stxn.TxnHash.String(),
//		Pubkey:    stxn.Pubkey.StringWithType(),
//		Signature: ToHex(stxn.Signature),
//		RawTxn:    ToHex(rawTxnByt),
//		BlockHash: "",
//	}, nil
//}
//
//func (t TxnScheme) toTxn() (*SignedTxn, error) {
//	rawTxn, err := DecodeUnsignedTxn(FromHex(t.RawTxn))
//	if err != nil {
//		return nil, err
//	}
//	pubkey, err := keypair.PubkeyFromStr(t.Pubkey)
//	if err != nil {
//		return nil, err
//	}
//	return &SignedTxn{
//		Raw:       rawTxn,
//		TxnHash:   HexToHash(t.TxnHash),
//		Pubkey:    pubkey,
//		Signature: FromHex(t.Signature),
//	}, nil
//}
//
//type EventScheme struct {
//	gorm.Model
//	Caller     string
//	BlockStage string
//	BlockHash  string
//	Height     BlockNum
//	TripodName string
//	ExecName   string
//	Value      string
//}
//
//func (EventScheme) TableName() string {
//	return "events"
//}
//
//func toEventScheme(event *Event) (EventScheme, error) {
//	return EventScheme{
//		Caller:     event.Caller.String(),
//		BlockStage: event.BlockStage,
//		BlockHash:  event.BlockHash.String(),
//		Height:     event.Height,
//		TripodName: event.TripodName,
//		ExecName:   event.FuncName,
//		Value:      event.Value,
//	}, nil
//}
//
//func (e EventScheme) toEvent() (*Event, error) {
//	return &Event{
//		Caller:      HexToAddress(e.Caller),
//		BlockStage:  e.BlockStage,
//		BlockHash:   HexToHash(e.BlockHash),
//		Height:      e.Height,
//		TripodName:  e.TripodName,
//		FuncName: e.ExecName,
//		Value:       e.Value,
//	}, nil
//
//}
//
//type ErrorScheme struct {
//	gorm.Model
//	Caller     string
//	BlockStage string
//	BlockHash  string
//	Height     BlockNum
//	TripodName string
//	ExecName   string
//	Error      string
//}
//
//func (ErrorScheme) TableName() string {
//	return "errors"
//}
//
//func toErrorScheme(err *Error) ErrorScheme {
//	return ErrorScheme{
//		Caller:     err.Caller.String(),
//		BlockStage: err.BlockStage,
//		BlockHash:  err.BlockHash.String(),
//		Height:     err.Height,
//		TripodName: err.TripodName,
//		ExecName:   err.FuncName,
//		Error:      err.Err,
//	}
//}
//
//func (e ErrorScheme) toError() *Error {
//	return &Error{
//		Caller:      HexToAddress(e.Caller),
//		BlockStage:  e.BlockStage,
//		BlockHash:   HexToHash(e.BlockHash),
//		Height:      e.Height,
//		TripodName:  e.TripodName,
//		FuncName: e.ExecName,
//		Err:         e.Error,
//	}
//}
