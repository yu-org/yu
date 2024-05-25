package tripod

import (
	. "github.com/yu-org/yu/core/types"
)

type DefaultTxnChecker struct{}

func (*DefaultTxnChecker) CheckTxn(*SignedTxn) error {
	return nil
}

type DefaultBlockVerifier struct{}

func (*DefaultBlockVerifier) VerifyBlock(*Block) bool {
	return true
}

type DefaultInit struct{}

func (*DefaultInit) InitChain(*Block) {}

type DefaultBlockStarter struct{}

func (*DefaultBlockStarter) StartBlock(*Block) {}

type DefaultBlockEnder struct{}

func (*DefaultBlockEnder) EndBlock(*Block) {}

type DefaultBlockFinalizer struct{}

func (*DefaultBlockFinalizer) FinalizeBlock(*Block) {}

type DefaultCommitter struct{}

func (*DefaultCommitter) Commit(*Block) {}
