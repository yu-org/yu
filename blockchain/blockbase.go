package blockchain

import (
	"errors"
	"gorm.io/gorm"
	. "yu/common"
	"yu/config"
	"yu/keypair"
	. "yu/result"
	ysql "yu/storage/sql"
	. "yu/txn"
)

type BlockBase struct {
	db ysql.SqlDB
}

func NewBlockBase(cfg *config.BlockBaseConf) (*BlockBase, error) {
	db, err := ysql.NewSqlDB(&cfg.BaseDB)
	if err != nil {
		return nil, err
	}

	err = db.CreateIfNotExist(&TxnScheme{})
	if err != nil {
		return nil, err
	}

	err = db.CreateIfNotExist(&EventScheme{})
	if err != nil {
		return nil, err
	}

	err = db.CreateIfNotExist(&ErrorScheme{})
	if err != nil {
		return nil, err
	}

	return &BlockBase{
		db: db,
	}, nil
}

func (bb *BlockBase) GetTxn(txnHash Hash) (*SignedTxn, error) {
	var ts TxnScheme
	bb.db.Db().Where(&TxnScheme{TxnHash: txnHash.String()}).First(&ts)
	return ts.toTxn()
}

func (bb *BlockBase) SetTxn(stxn *SignedTxn) error {
	txnSm, err := toTxnScheme(stxn)
	if err != nil {
		return err
	}
	bb.db.Db().Create(&txnSm)
	return nil
}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]*SignedTxn, error) {
	var tss []TxnScheme
	bb.db.Db().Where(&TxnScheme{BlockHash: blockHash.String()}).Find(&tss)
	itxns := make([]*SignedTxn, 0)
	for _, ts := range tss {
		stxn, err := ts.toTxn()
		if err != nil {
			return nil, err
		}
		itxns = append(itxns, stxn)
	}
	return itxns, nil
}

func (bb *BlockBase) SetTxns(blockHash Hash, txns []*SignedTxn) error {
	txnSms := make([]TxnScheme, 0)
	for _, stxn := range txns {
		txnSm, err := newTxnScheme(blockHash, stxn)
		if err != nil {
			return err
		}
		txnSms = append(txnSms, txnSm)
	}
	if len(txnSms) > 0 {
		bb.db.Db().Create(&txnSms)
	}
	return nil
}

func (bb *BlockBase) GetEvents(blockHash Hash) ([]*Event, error) {
	var ess []EventScheme
	bb.db.Db().Where(&EventScheme{BlockHash: blockHash.String()}).Find(&ess)
	events := make([]*Event, 0)
	for _, es := range ess {
		e, err := es.toEvent()
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (bb *BlockBase) SetEvents(events []*Event) error {
	eventSms := make([]EventScheme, 0)
	for _, event := range events {
		eventSm, err := toEventScheme(event)
		if err != nil {
			return err
		}
		eventSms = append(eventSms, eventSm)
	}
	if len(eventSms) > 0 {
		bb.db.Db().Create(&eventSms)
	}
	return nil
}

func (bb *BlockBase) GetErrors(blockHash Hash) ([]*Error, error) {
	var ess []ErrorScheme
	bb.db.Db().Where(&ErrorScheme{BlockHash: blockHash.String()}).Find(&ess)
	errs := make([]*Error, 0)
	for _, es := range ess {
		errs = append(errs, es.toError())
	}
	return errs, nil
}

func (bb *BlockBase) SetErrors(errs []*Error) error {
	errSms := make([]ErrorScheme, 0)
	for _, err := range errs {
		errSms = append(errSms, toErrorScheme(err))
	}
	if len(errSms) > 0 {
		bb.db.Db().Create(&errSms)
	}
	return nil
}

type TxnScheme struct {
	TxnHash   string `gorm:"primaryKey"`
	Pubkey    string
	KeyType   string
	Signature string
	RawTxn    string

	BlockHash string
}

func (TxnScheme) TableName() string {
	return "txns"
}

func newTxnScheme(blockHash Hash, stxn *SignedTxn) (TxnScheme, error) {
	txnSm, err := toTxnScheme(stxn)
	if err != nil {
		return TxnScheme{}, err
	}
	txnSm.BlockHash = blockHash.String()
	return txnSm, nil
}

func toTxnScheme(stxn *SignedTxn) (TxnScheme, error) {
	rawTxnByt, err := stxn.GetRaw().Encode()
	if err != nil {
		return TxnScheme{}, err
	}
	return TxnScheme{
		TxnHash:   stxn.GetTxnHash().String(),
		Pubkey:    stxn.GetPubkey().String(),
		KeyType:   stxn.GetPubkey().Type(),
		Signature: ToHex(stxn.GetSignature()),
		RawTxn:    ToHex(rawTxnByt),
		BlockHash: "",
	}, nil
}

func (t TxnScheme) toTxn() (*SignedTxn, error) {
	ut := &UnsignedTxn{}
	rawTxn, err := ut.Decode(FromHex(t.RawTxn))
	if err != nil {
		return nil, err
	}
	pubkey, err := keypair.PubKeyFromBytes(t.KeyType, FromHex(t.Pubkey))
	if err != nil {
		return nil, err
	}
	return &SignedTxn{
		Raw:       rawTxn,
		TxnHash:   HexToHash(t.TxnHash),
		Pubkey:    pubkey,
		Signature: FromHex(t.Signature),
	}, nil
}

type EventScheme struct {
	gorm.Model
	Caller     string
	BlockStage string
	BlockHash  string
	Height     BlockNum
	TripodName string
	ExecName   string
	Value      string
}

func (EventScheme) TableName() string {
	return "events"
}

func toEventScheme(event *Event) (EventScheme, error) {
	return EventScheme{
		Caller:     event.Caller.String(),
		BlockStage: event.BlockStage,
		BlockHash:  event.BlockHash.String(),
		Height:     event.Height,
		TripodName: event.TripodName,
		ExecName:   event.ExecName,
		Value:      event.Value,
	}, nil
}

func (e EventScheme) toEvent() (*Event, error) {
	return &Event{
		Caller:     HexToAddress(e.Caller),
		BlockStage: e.BlockStage,
		BlockHash:  HexToHash(e.BlockHash),
		Height:     e.Height,
		TripodName: e.TripodName,
		ExecName:   e.ExecName,
		Value:      e.Value,
	}, nil

}

type ErrorScheme struct {
	gorm.Model
	Caller     string
	BlockStage string
	BlockHash  string
	Height     BlockNum
	TripodName string
	ExecName   string
	Error      string
}

func (ErrorScheme) TableName() string {
	return "errors"
}

func toErrorScheme(err *Error) ErrorScheme {
	return ErrorScheme{
		Caller:     err.Caller.String(),
		BlockStage: err.BlockStage,
		BlockHash:  err.BlockHash.String(),
		Height:     err.Height,
		TripodName: err.TripodName,
		ExecName:   err.ExecName,
		Error:      err.Err.Error(),
	}
}

func (e ErrorScheme) toError() *Error {
	return &Error{
		Caller:     HexToAddress(e.Caller),
		BlockStage: e.BlockStage,
		BlockHash:  HexToHash(e.BlockHash),
		Height:     e.Height,
		TripodName: e.TripodName,
		ExecName:   e.ExecName,
		Err:        errors.New(e.Error),
	}
}
