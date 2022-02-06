package txpool

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	ytime "github.com/yu-org/yu/utils/time"
	"sync"
)

type TxPool struct {
	sync.RWMutex

	poolSize   uint64
	TxnMaxSize int
	startTS    uint64

	unpackedTxns *orderedTxns
	yudb         IyuDB

	baseChecks   []TxnCheckFn
	tripodChecks []TxnCheckFn
}

func NewTxPool(cfg *TxpoolConf, base IyuDB) *TxPool {
	ordered := newOrderedTxns()

	tp := &TxPool{
		poolSize:     cfg.PoolSize,
		TxnMaxSize:   cfg.TxnMaxSize,
		unpackedTxns: ordered,
		startTS:      ytime.NowNanoTsU64(),
		yudb:         base,
		baseChecks:   make([]TxnCheckFn, 0),
		tripodChecks: make([]TxnCheckFn, 0),
	}
	allUnpacked, err := tp.yudb.GetAllUnpackedTxns()
	if err != nil {
		logrus.Fatal("get all unpacked txns from txpool db failed: ", err)
	}
	for _, tx := range allUnpacked {
		tp.unpackedTxns.insert(tx)
	}
	return tp
}

func WithDefaultChecks(cfg *TxpoolConf, base IyuDB) *TxPool {
	tp := NewTxPool(cfg, base)
	return tp.withDefaultBaseChecks()
}

func (tp *TxPool) withDefaultBaseChecks() *TxPool {
	tp.baseChecks = []TxnCheckFn{
		tp.checkPoolLimit,
		tp.checkTxnSize,
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

func (tp *TxPool) WithBaseCheck(checkFn TxnCheckFn) ItxPool {
	tp.baseChecks = append(tp.baseChecks, checkFn)
	return tp
}

func (tp *TxPool) WithTripodCheck(tri TxnCheckTripod) ItxPool {
	tp.tripodChecks = append(tp.tripodChecks, tri.CheckTxn)
	return tp
}

func (tp *TxPool) Insert(stxn *SignedTxn) error {
	errs := tp.BatchInsert(FromArray(stxn))
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

func (tp *TxPool) BatchInsert(txns SignedTxns) []error {
	tp.Lock()
	defer tp.Unlock()
	errs := make([]error, 0)
	for _, stxn := range txns {
		if tp.unpackedTxns.exist(stxn) {
			continue
		}
		// check replay attack
		if tp.yudb.ExistTxn(stxn.TxnHash) {
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
		err = tp.yudb.SetTxn(stxn)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		tp.unpackedTxns.insert(stxn)
	}
	return errs
}

func (tp *TxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	tp.RLock()
	defer tp.RUnlock()
	return tp.yudb.GetTxn(hash)
}

func (tp *TxPool) Pack(numLimit uint64) ([]*SignedTxn, error) {
	return tp.PackFor(numLimit, func(*SignedTxn) bool {
		return true
	})
}

func (tp *TxPool) PackFor(numLimit uint64, filter func(txn *SignedTxn) bool) ([]*SignedTxn, error) {
	tp.Lock()
	defer tp.Unlock()
	txns := tp.unpackedTxns.gets(numLimit, filter)
	return txns, nil
}

func (tp *TxPool) Reset(block *CompactBlock) error {
	tp.Lock()
	defer tp.Unlock()
	err := tp.yudb.Packs(block.Hash, block.TxnsHashes)
	if err != nil {
		return err
	}
	tp.unpackedTxns.deletes(block.TxnsHashes)
	return nil
}

// ------------------- check txn rules ----------------------

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

	return tp.TripodsCheck(stxn)
}

func (tp *TxPool) checkPoolLimit(*SignedTxn) error {
	if uint64(tp.unpackedTxns.len()) >= tp.poolSize {
		return PoolOverflow
	}
	return nil
}

func (tp *TxPool) checkTxnSize(stxn *SignedTxn) error {
	if stxn.Size() > tp.TxnMaxSize {
		return TxnTooLarge
	}
	return nil
}

type TxnCheckFn func(*SignedTxn) error

type TxnCheckTripod interface {
	CheckTxn(*SignedTxn) error
}

func Check(checks []TxnCheckFn, stxn *SignedTxn) error {
	for _, check := range checks {
		err := check(stxn)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckSignature(stxn *SignedTxn) error {
	sig := stxn.Signature
	ecall := stxn.Raw.Ecall
	if !stxn.Pubkey.VerifySignature(ecall.Bytes(), sig) {
		return TxnSignatureErr
	}
	return nil
}
