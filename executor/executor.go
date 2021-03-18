package executor

type IExecutor interface {
	Stash() error
	Execute() error
	Commit() error
	Rollback() error
}
