package txdb

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/types"
	"math/big"
	"testing"
)

var yudbCfg = &config.TxDBConf{BaseDB: config.SqlDbConf{
	SqlDbType: "sqlite",
	Dsn:       "./test_yudb",
}}

var (
	pub, _           = keypair.GenEdKeyWithSecret([]byte("yu"))
	addr             = pub.Address()
	blockHash        = common.BigToHash(big.NewInt(10))
	txn1, txn2, txn3 *types.SignedTxn
)

func init() {
	var err error
	txn1, err = types.NewSignedTxn(addr, &common.WrCall{
		TripodName: "tripod-1",
	}, pub, nil)
	if err != nil {
		panic(err)
	}
	txn2, err = types.NewSignedTxn(addr, &common.WrCall{
		TripodName: "tripod-2",
	}, pub, nil)
	if err != nil {
		panic(err)
	}
	txn3, err = types.NewSignedTxn(addr, &common.WrCall{
		TripodName: "tripod-3",
	}, pub, nil)
	if err != nil {
		panic(err)
	}
}

func TestExistTxn(t *testing.T) {
	yudb := NewTxDB(yudbCfg)
	err := yudb.SetTxn(txn1)
	if err != nil {
		panic(err)
	}
	assert.True(t, yudb.ExistTxn(txn1.TxnHash))
	assert.True(t, !yudb.ExistTxn(txn2.TxnHash))
}

func TestPacks(t *testing.T) {
	yudb := NewTxDB(yudbCfg)
	insertTxns(yudb)
	err := yudb.Packs(blockHash, []common.Hash{txn1.TxnHash, txn2.TxnHash})
	if err != nil {
		panic(err)
	}
	unpacks, err := yudb.GetAllUnpackedTxns()
	if err != nil {
		panic(err)
	}
	assert.Equal(t, txn3.TxnHash.String(), unpacks[0].TxnHash.String())
}

func insertTxns(yudb *TxDB) {
	err := yudb.SetTxn(txn1)
	if err != nil {
		panic(err)
	}
	println("insert txn1: ", txn1.TxnHash.String())

	err = yudb.SetTxn(txn2)
	if err != nil {
		panic(err)
	}
	println("insert txn2: ", txn2.TxnHash.String())

	err = yudb.SetTxn(txn3)
	if err != nil {
		panic(err)
	}
	println("insert txn3: ", txn3.TxnHash.String())
}
