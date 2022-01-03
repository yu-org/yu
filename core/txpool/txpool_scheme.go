package txpool

//
//import (
//	. "github.com/yu-org/yu/common"
//	. "github.com/yu-org/yu/core/types"
//	"gorm.io/gorm"
//)
//
//type TxpoolScheme struct {
//	TxnHash  string `gorm:"primaryKey"`
//	Txn      string
//	IsPacked bool
//}
//
//func (TxpoolScheme) TableName() string {
//	return "txpool"
//}
//
//func (tp *TxPool) insertToDB(txn *SignedTxn) error {
//	byt, err := txn.Encode()
//	if err != nil {
//		return err
//	}
//	scheme := &TxpoolScheme{
//		TxnHash:  txn.TxnHash.String(),
//		Txn:      ToHex(byt),
//		IsPacked: false,
//	}
//	return tp.db.Db().Create(scheme).Error
//}
//
//func (tp *TxPool) getUnpacked(hash Hash) (*SignedTxn, error) {
//	var scheme TxpoolScheme
//	tp.db.Db().Where(&TxpoolScheme{
//		TxnHash:  hash.String(),
//		IsPacked: false,
//	}).First(&scheme)
//	return DecodeSignedTxn(FromHex(scheme.Txn))
//}
//
//func (tp *TxPool) getAllUnpacked() (txns []*SignedTxn, err error) {
//	var schemes []*TxpoolScheme
//	err = tp.db.Db().Where(&TxpoolScheme{IsPacked: false}).Find(&schemes).Error
//	if err != nil {
//		return
//	}
//	for _, scheme := range schemes {
//		var txn *SignedTxn
//		txn, err = DecodeSignedTxn(FromHex(scheme.Txn))
//		if err != nil {
//			return
//		}
//		txns = append(txns, txn)
//	}
//	return
//}
//
//func (tp *TxPool) packByHashes(hashes []Hash) error {
//	return tp.db.Db().Transaction(func(tx *gorm.DB) error {
//		for _, hash := range hashes {
//			tx.Where(&TxpoolScheme{TxnHash: hash.String()}).Updates(TxpoolScheme{IsPacked: true})
//		}
//		return nil
//	})
//}
//
//func (tp *TxPool) cleanPacked() error {
//	return tp.db.Db().Where(&TxpoolScheme{IsPacked: true}).Delete(TxpoolScheme{}).Error
//}
