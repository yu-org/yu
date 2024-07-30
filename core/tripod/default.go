package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type DefaultTxnChecker struct{}

func (*DefaultTxnChecker) CheckTxn(*SignedTxn) error {
	return nil
}

type DefaultBlockVerifier struct{}

func (*DefaultBlockVerifier) VerifyBlock(*Block) error {
	return nil
}

type DefaultInit struct{}

func (*DefaultInit) InitChain(*Block) {}

type DefaultBlockCycle struct{}

func (*DefaultBlockCycle) StartBlock(*Block)    {}
func (*DefaultBlockCycle) EndBlock(*Block)      {}
func (*DefaultBlockCycle) FinalizeBlock(*Block) {}

type DefaultPreTxnHandler struct{}

func (*DefaultPreTxnHandler) PreHandleTxn(*SignedTxn) error {
	return nil
}

type DefaultCommitter struct{}

func (*DefaultCommitter) Commit(*Block) {}
