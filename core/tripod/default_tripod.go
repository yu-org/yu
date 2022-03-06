package tripod

import (
	. "github.com/yu-org/yu/core/chain_env"
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
)

type DefaultTripod struct {
	*TripodMeta
	*ChainEnv

	txnChecker    TxnCheckFn
	blockVerifier BlockVerifier
}

type BlockVerifier func(block *CompactBlock) bool

func NewDefaultTripod(name string) *DefaultTripod {
	meta := NewTripodMeta(name)
	return &DefaultTripod{
		TripodMeta: meta,
	}
}

func (dt *DefaultTripod) SetTxnChecker(checker TxnCheckFn) {
	dt.txnChecker = checker
}

func (dt *DefaultTripod) SetBlockVerifier(fn BlockVerifier) {
	dt.blockVerifier = fn
}

func (dt *DefaultTripod) GetTripodMeta() *TripodMeta {
	return dt.TripodMeta
}

func (dt *DefaultTripod) SetChainEnv(env *ChainEnv) {
	dt.ChainEnv = env
}

func (dt *DefaultTripod) CheckTxn(txn *SignedTxn) error {
	if dt.txnChecker == nil {
		return nil
	}
	return dt.txnChecker(txn)
}

func (dt *DefaultTripod) VerifyBlock(block *CompactBlock) bool {
	if dt.blockVerifier == nil {
		return true
	}
	return dt.blockVerifier(block)
}

func (*DefaultTripod) InitChain() error {
	return nil
}

func (*DefaultTripod) StartBlock(*CompactBlock) error {
	return nil
}

func (*DefaultTripod) EndBlock(*CompactBlock) error {
	return nil
}

func (*DefaultTripod) FinalizeBlock(*CompactBlock) error {
	return nil
}
