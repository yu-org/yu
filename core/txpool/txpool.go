package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	. "github.com/yu-org/yu/infra/storage/kv"
	ytime "github.com/yu-org/yu/utils/time"
	"sync"
)

// This implementation only use for Local-Node mode.
type TxPool struct {
	sync.RWMutex

	poolSize   uint64
	TxnMaxSize int

	Txns         []*TxpoolScheme
	startPackIdx int

	startTS uint64
	db      KV

	baseChecks   []TxnCheck
	tripodChecks []TxnCheck
}

func NewTxPool(cfg *TxpoolConf) *TxPool {
	db, err := NewKV(&cfg.DB)
	if err != nil {
		logrus.Fatal("init txpool error: ", err)
	}
	return &TxPool{
		poolSize:     cfg.PoolSize,
		TxnMaxSize:   cfg.TxnMaxSize,
		Txns:         make([]*TxpoolScheme, 0),
		startPackIdx: 0,
		startTS:      ytime.NowNanoTsU64(),
		db:           db,
		baseChecks:   make([]TxnCheck, 0),
		tripodChecks: make([]TxnCheck, 0),
	}
}

func LocalWithDefaultChecks(cfg *TxpoolConf) *TxPool {
	tp := NewTxPool(cfg)
	return tp.withDefaultBaseChecks()
}

func (tp *TxPool) withDefaultBaseChecks() *TxPool {
	tp.baseChecks = []TxnCheck{
		tp.checkPoolLimit,
		tp.checkTxnSize,
		tp.checkTimeout,
		tp.checkSignature,
	}
	return tp
}

func (tp *TxPool) NewEmptySignedTxn() *SignedTxn {
	return &SignedTxn{}
}

func (tp *TxPool) NewEmptySignedTxns() SignedTxns {
	return make([]*SignedTxn, 0)
}

func (tp *TxPool) PoolSize() uint64 {
	return tp.poolSize
}

func (tp *TxPool) WithBaseChecks(checkFns []TxnCheck) ItxPool {
	tp.baseChecks = append(tp.baseChecks, checkFns...)
	return tp
}

func (tp *TxPool) WithTripodChecks(checkFns []TxnCheck) ItxPool {
	tp.tripodChecks = append(tp.tripodChecks, checkFns...)
	return tp
}

// insert into txpool
func (tp *TxPool) Insert(stxn *SignedTxn) error {
	errs := tp.BatchInsert(FromArray(stxn))
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

// batch insert into txpool
func (tp *TxPool) BatchInsert(txns SignedTxns) []error {
	tp.Lock()
	defer tp.Unlock()
	errs := make([]error, 0)
	for _, stxn := range txns {

		if tp.db.Exist(stxn.TxnHash.Bytes()) {
			continue
		}
		err := tp.BaseCheck(stxn)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = tp.TripodsCheck(stxn)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = tp.insertTx(stxn)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	return errs
}

// package some txns to send to tripods
func (tp *TxPool) Pack(numLimit uint64) ([]*SignedTxn, error) {
	return tp.PackFor(numLimit, func(*SignedTxn) error {
		return nil
	})
}

func (tp *TxPool) PackFor(numLimit uint64, filter func(*SignedTxn) error) ([]*SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	stxns := make([]*SignedTxn, 0)
	for i := 0; i < int(numLimit); i++ {
		if i >= len(tp.Txns) {
			break
		}
		logrus.Debug("********************** pack txn: ", tp.Txns[i].TxnHash.String())
		err := filter(tp.Txns[i])
		if err != nil {
			return nil, err
		}
		stxns = append(stxns, tp.Txns[i])
		tp.startPackIdx++
	}
	return stxns, nil
}

func (tp *TxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	return tp.getTx(hash)
}

func (tp *TxPool) Packed(hashes []Hash) error {
	tp.Lock()
	for _, hash := range hashes {
		var idx int
		idx, tp.Txns = tp.Txns.Remove(hash)
		if idx == -1 {
			continue
		}
		tp.db.Delete(hash.Bytes())

		if idx < tp.startPackIdx {
			tp.startPackIdx--
		}
	}
	tp.Unlock()
	return nil
}

// remove txns after execute all tripods
func (tp *TxPool) Reset() error {
	tp.Lock()
	for _, stxn := range tp.Txns[:tp.startPackIdx] {
		tp.db.Delete(stxn.TxnHash.Bytes())
	}
	tp.Txns = tp.Txns[tp.startPackIdx:]
	tp.startPackIdx = 0

	tp.startTS = ytime.NowNanoTsU64()
	tp.Unlock()
	return nil
}

func (tp *TxPool) BaseCheck(stxn *SignedTxn) error {
	return Check(tp.baseChecks, stxn)
}

func (tp *TxPool) TripodsCheck(stxn *SignedTxn) error {
	return Check(tp.tripodChecks, stxn)
}

func (tp *TxPool) NecessaryCheck(stxn *SignedTxn) (err error) {
	err = tp.checkTxnSize(stxn)
	if err != nil {
		return
	}
	err = tp.checkSignature(stxn)
	if err != nil {
		return
	}

	return tp.TripodsCheck(stxn)
}

func (tp *TxPool) checkPoolLimit(*SignedTxn) error {
	if uint64(len(tp.Txns)) >= tp.poolSize {
		return PoolOverflow
	}
	return nil
}

func (tp *TxPool) checkTimeout(stxn *SignedTxn) error {
	if stxn.Raw.Timestamp < tp.startTS {
		return TxnTimeoutErr
	}
	return nil
}

func (tp *TxPool) checkSignature(stxn *SignedTxn) error {
	sig := stxn.Signature
	ecall := stxn.Raw.Ecall
	if !stxn.Pubkey.VerifySignature(ecall.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}

func (tp *TxPool) checkTxnSize(stxn *SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return nil
}

func (tp *TxPool) insertTx(txn *SignedTxn) error {
	itxn := &TxpoolScheme{
		Txn:      txn,
		IsPacked: false,
	}
	byt, err := itxn.encode()
	if err != nil {
		return err
	}
	err = tp.db.Set(txn.TxnHash.Bytes(), byt)
	if err != nil {
		return err
	}
	tp.Txns = append(tp.Txns, itxn)
	return nil
}

func (tp *TxPool) getTx(hash Hash) (*SignedTxn, error) {
	byt, err := tp.db.Get(hash.Bytes())
	if err != nil {
		return nil, err
	}
	itxn, err := new(TxpoolScheme).decode(byt)
	if err != nil {
		return nil, err
	}
	return itxn.Txn, nil
}

func (tp *TxPool) cleanPacked() error {

}

func (tp *TxPool) pack(hash Hash) {

}

type TxpoolScheme struct {
	Txn      *SignedTxn
	IsPacked bool
}

func boolToByte(b bool) byte {
	if b {
		return 0
	}
	return 1
}

func byteToBool(b byte) bool {
	return b == 0
}

func (it *TxpoolScheme) encode() (byt []byte, err error) {
	byt, err = it.Txn.Encode()
	if err != nil {
		return
	}
	byt = append(byt, boolToByte(it.IsPacked))
	return
}

func (it *TxpoolScheme) decode(data []byte) (*TxpoolScheme, error) {
	txn, err := DecodeSignedTxn(data[:len(data)-1])
	if err != nil {
		return nil, err
	}
	it.Txn = txn
	it.IsPacked = byteToBool(data[len(data)-1])
	return it, nil
}

type TxnCheck func(*SignedTxn) error

func Check(checks []TxnCheck, stxn *SignedTxn) error {
	for _, check := range checks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}
