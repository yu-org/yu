package txdb

import (
	"database/sql"
	"sync"

	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"

	csql "github.com/yu-org/yu/infra/storage/sql"
)

type sqliteDbConn struct {
	sync.Mutex
	db *sql.DB
}

type txnSqliteStorage struct {
	txnConn *sqliteDbConn
}

func (ts *txnSqliteStorage) initdb() error {
	db, err := csql.GlobalSqliteManager.CreateStore("txn.db")
	if err != nil {
		return err
	}
	ts.txnConn = &sqliteDbConn{db: db}
	return ts.applyScheme()
}

func (ts *txnSqliteStorage) applyScheme() error {
	ts.txnConn.Lock()
	defer ts.txnConn.Unlock()
	_, err := ts.txnConn.db.Exec(`CREATE TABLE IF NOT EXISTS txn (key BLOB PRIMARY KEY, value BLOB, createdTimestap INTEGER DEFAULT (STRFTIME('%s', 'NOW')));`)
	return err
}

func (ts *txnSqliteStorage) SetTxns(txns []*SignedTxn) (err error) {
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	for _, txn := range txns {
		v, err := txn.Encode()
		if err != nil {
			return err
		}
		keys = append(keys, txn.TxnHash.Bytes())
		values = append(values, v)
	}

	ts.txnConn.Lock()
	defer ts.txnConn.Unlock()
	tx, err := ts.txnConn.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into txn (key, value) values(?, ?)")
	if err != nil {
		return err
	}
	for index, key := range keys {
		_, err := stmt.Exec(key, values[index])
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (ts *txnSqliteStorage) ExistTxn(txnHash Hash) bool {
	h := txnHash.Bytes()
	ts.txnConn.Lock()
	defer ts.txnConn.Unlock()
	db := ts.txnConn.db
	stmt, err := db.Prepare("select count(*) from txn where key=?")
	if err != nil {
		return false
	}
	row, err := stmt.Query(h)
	if err != nil {
		return false
	}
	defer row.Close()
	var cnt int
	for row.Next() {
		row.Scan(&cnt)
	}
	return cnt > 0
}

func (ts *txnSqliteStorage) GetTxn(txnHash Hash) (txn *SignedTxn, err error) {
	value, err := ts.getTxn(txnHash.Bytes())
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	txn, err = DecodeSignedTxn(value)
	return txn, err
}

func (ts *txnSqliteStorage) getTxn(txnHash []byte) (value []byte, err error) {
	ts.txnConn.Lock()
	defer ts.txnConn.Unlock()
	db := ts.txnConn.db
	stmt, err := db.Prepare("select value from txn where key=?")
	if err != nil {
		return nil, err
	}
	r, err := stmt.Query(txnHash)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var v []byte
	for r.Next() {
		err = r.Scan(&v)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

type receiptSqliteStorage struct {
	conn *sqliteDbConn
}

func (rs *receiptSqliteStorage) initdb() error {
	db, err := csql.GlobalSqliteManager.CreateStore("receipt.db")
	if err != nil {
		return err
	}
	rs.conn = &sqliteDbConn{db: db}
	return rs.applyScheme()
}

func (rs *receiptSqliteStorage) applyScheme() error {
	rs.conn.Lock()
	defer rs.conn.Unlock()
	_, err := rs.conn.db.Exec(`CREATE TABLE IF NOT EXISTS receipt (key BLOB PRIMARY KEY, value BLOB, createdTimestap INTEGER DEFAULT (STRFTIME('%s', 'NOW')));`)
	return err
}

func (rs *receiptSqliteStorage) SetReceipts(receipts map[Hash]*Receipt) error {
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	for key, receipt := range receipts {
		v, err := receipt.Encode()
		if err != nil {
			return err
		}
		keys = append(keys, key.Bytes())
		values = append(values, v)
	}
	rs.conn.Lock()
	defer rs.conn.Unlock()
	tx, err := rs.conn.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("insert into receipt (key, value) values(?, ?)")
	if err != nil {
		return err
	}
	for index, key := range keys {
		_, err := stmt.Exec(key, values[index])
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (rs *receiptSqliteStorage) SetReceipt(txHash Hash, receipt *Receipt) error {
	key := txHash.Bytes()
	value, err := receipt.Encode()
	if err != nil {
		return err
	}
	rs.conn.Lock()
	defer rs.conn.Unlock()
	db := rs.conn.db
	stmt, err := db.Prepare("insert into receipt (key, value) values(?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(key, value)
	return err
}

func (rs *receiptSqliteStorage) GetReceipt(txHash Hash) (*Receipt, error) {
	rs.conn.Lock()
	defer rs.conn.Unlock()
	db := rs.conn.db
	stmt, err := db.Prepare("select value from receipt where key=?")
	if err != nil {
		return nil, err
	}
	row, err := stmt.Query(txHash.Bytes())
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var v []byte
	for row.Next() {
		err = row.Scan(&v)
		if err != nil {
			return nil, err
		}
	}
	if v == nil {
		return nil, nil
	}
	receipt := new(Receipt)
	err = receipt.Decode(v)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}
