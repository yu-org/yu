package txpool

import (
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
	txdb         ItxDB

	baseChecks   []TxnCheckFn
	tripodChecks []TxnCheckFn
}

func NewTxPool(cfg *TxpoolConf, base ItxDB) *TxPool {
	ordered := newOrderedTxns()

	tp := &TxPool{
		poolSize:     cfg.PoolSize,
		TxnMaxSize:   cfg.TxnMaxSize,
		unpackedTxns: ordered,
		startTS:      ytime.NowNanoTsU64(),
		txdb:         base,
		baseChecks:   make([]TxnCheckFn, 0),
		tripodChecks: make([]TxnCheckFn, 0),
	}
	return tp
}

func WithDefaultChecks(cfg *TxpoolConf, base ItxDB) *TxPool {
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

func (tp *TxPool) Exist(stxn *SignedTxn) bool {
	tp.RLock()
	defer tp.RUnlock()
	// check replay attack
	return tp.txdb.ExistTxn(stxn.TxnHash)
}

func (tp *TxPool) CheckTxn(stxn *SignedTxn) (err error) {
	err = tp.BaseCheck(stxn)
	if err != nil {
		return
	}
	return tp.TripodsCheck(stxn)
}

func (tp *TxPool) Insert(stxn *SignedTxn) error {
	tp.Lock()
	defer tp.Unlock()
	tp.unpackedTxns.insert(stxn)
	return nil
}

func (tp *TxPool) GetTxn(hash Hash) (*SignedTxn, error) {
	tp.RLock()
	defer tp.RUnlock()
	return tp.unpackedTxns.get(hash), nil
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

func (tp *TxPool) Reset(txns SignedTxns) error {
	tp.Lock()
	defer tp.Unlock()
	tp.unpackedTxns.deletes(txns.Hashes())
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
	if uint64(tp.unpackedTxns.size()) >= tp.poolSize {
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
	hash, err := ecall.Hash()
	if err != nil {
		return err
	}
	if !stxn.Pubkey.VerifySignature(hash, sig) {
		return TxnSignatureErr
	}
	return nil
}
