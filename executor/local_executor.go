package executor

import . "yu/txn"

type LocalExecutor struct {
	UnExecuteTxns []IsignedTxn
}

func (le *LocalExecutor) Stash() error {

}

func (le *LocalExecutor) Execute() error {

}

func (le *LocalExecutor) Commit() error {

}

func (le *LocalExecutor) Rollback() error {

}
