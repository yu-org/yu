package blockbase

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/result"
	"github.com/yu-org/yu/core/types"
	ysql "github.com/yu-org/yu/infra/storage/sql"
)

type BlockBase struct {
	db ysql.SqlDB
}

func NewBlockBase(cfg *config.BlockBaseConf) *BlockBase {
	db, err := ysql.NewSqlDB(&cfg.BaseDB)
	if err != nil {
		logrus.Fatal("init blockbase SQL db error: ", err)
	}

	err = db.CreateIfNotExist(&TxnScheme{})
	if err != nil {
		logrus.Fatal("create blockbase TXN sceme error: ", err)
	}

	err = db.CreateIfNotExist(&EventScheme{})
	if err != nil {
		logrus.Fatal("create blockbase Event sceme error: ", err)
	}

	err = db.CreateIfNotExist(&ErrorScheme{})
	if err != nil {
		logrus.Fatal("create blockbase Error sceme error: ", err)
	}

	return &BlockBase{
		db: db,
	}
}

func (bb *BlockBase) GetTxn(txnHash Hash) (*types.SignedTxn, error) {
	var ts TxnScheme
	bb.db.Db().Where(&TxnScheme{TxnHash: txnHash.String()}).First(&ts)
	return ts.toTxn()
}

func (bb *BlockBase) SetTxn(stxn *types.SignedTxn) error {
	txnSm, err := toTxnScheme(stxn)
	if err != nil {
		return err
	}
	bb.db.Db().Create(&txnSm)
	return nil
}

func (bb *BlockBase) GetTxns(blockHash Hash) ([]*types.SignedTxn, error) {
	var tss []TxnScheme
	bb.db.Db().Where(&TxnScheme{BlockHash: blockHash.String()}).Find(&tss)
	itxns := make([]*types.SignedTxn, 0)
	for _, ts := range tss {
		stxn, err := ts.toTxn()
		if err != nil {
			return nil, err
		}
		itxns = append(itxns, stxn)
	}
	return itxns, nil
}

func (bb *BlockBase) SetTxns(blockHash Hash, txns []*types.SignedTxn) error {
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

func (bb *BlockBase) SetError(err *Error) error {
	if err == nil {
		return nil
	}
	errscm := toErrorScheme(err)
	bb.db.Db().Create(&errscm)
	return nil
}
