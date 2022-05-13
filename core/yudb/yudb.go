package yudb

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	ysql "github.com/yu-org/yu/infra/storage/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type YuDB struct {
	db ysql.SqlDB
}

func NewYuDB(cfg *config.YuDBConf) *YuDB {
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

	return &YuDB{
		db: db,
	}
}

func (bb *YuDB) GetTxn(txnHash Hash) (*SignedTxn, error) {
	var ts TxnScheme
	bb.db.Db().Where(TxnScheme{TxnHash: txnHash.String()}).Find(&ts)
	return ts.toTxn()
}

func (bb *YuDB) ExistTxn(txnHash Hash) bool {
	var ts TxnScheme
	result := bb.db.Db().Debug().Where(TxnScheme{TxnHash: txnHash.String()}).Find(&ts)
	return result.RowsAffected > 0
}

func (bb *YuDB) SetTxn(stxn *SignedTxn) error {
	txnSm, err := toTxnScheme(stxn)
	if err != nil {
		return err
	}
	bb.db.Db().Create(&txnSm)
	return nil
}

func (bb *YuDB) GetAllUnpackedTxns() (txns []*SignedTxn, err error) {
	var schemes []*TxnScheme
	err = bb.db.Db().Where(&TxnScheme{IsPacked: false}).Find(&schemes).Error
	if err != nil {
		return
	}
	for _, scheme := range schemes {
		var txn *SignedTxn
		txn, err = scheme.toTxn()
		if err != nil {
			return
		}
		txns = append(txns, txn)
	}
	return
}

func (bb *YuDB) Packs(block Hash, txns []Hash) error {
	return bb.db.Db().Transaction(func(tx *gorm.DB) error {
		for _, txn := range txns {
			err := tx.Where(TxnScheme{TxnHash: txn.String()}).
				Updates(TxnScheme{
					BlockHash: block.String(),
					IsPacked:  true,
				}).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (bb *YuDB) Pack(block, txn Hash) error {
	return bb.db.Db().Where(TxnScheme{TxnHash: txn.String()}).
		Updates(TxnScheme{
			BlockHash: block.String(),
			IsPacked:  true,
		}).Error
}

func (bb *YuDB) GetTxns(blockHash Hash) ([]*SignedTxn, error) {
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

func (bb *YuDB) SetTxns(blockHash Hash, txns []*SignedTxn) error {
	txnSms := make([]TxnScheme, 0)
	for _, stxn := range txns {
		txnSm, err := newTxnScheme(blockHash, stxn)
		if err != nil {
			return err
		}
		txnSm.IsPacked = true
		txnSms = append(txnSms, txnSm)
	}

	if len(txnSms) > 0 {
		bb.db.Db().Clauses(clause.OnConflict{UpdateAll: true}).Create(&txnSms)
	}
	return nil
}

func (bb *YuDB) GetEvents(blockHash Hash) ([]*Event, error) {
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

func (bb *YuDB) SetEvents(events []*Event) error {
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

func (bb *YuDB) GetErrors(blockHash Hash) ([]*Error, error) {
	var ess []ErrorScheme
	bb.db.Db().Where(&ErrorScheme{BlockHash: blockHash.String()}).Find(&ess)
	errs := make([]*Error, 0)
	for _, es := range ess {
		errs = append(errs, es.toError())
	}
	return errs, nil
}

func (bb *YuDB) SetError(err *Error) error {
	if err == nil {
		return nil
	}
	errscm := toErrorScheme(err)
	bb.db.Db().Create(&errscm)
	return nil
}
