package yerror

import "github.com/pkg/errors"

var TypeErr = errors.New("the type of params error")

var (
	TripodNotFound = errors.New("Tripod NOT Found")
	ExecNotFound   = errors.New("Execution NOT Found")
	QryNotFound    = errors.New("Query NOT Found")
	WorkerDead     = errors.New("Worker Dead")
)

var NoKvdbType = errors.New("no kvdb type")

var (
	PoolOverflow    error = errors.New("pool size is full")
	TxnSignatureErr error = errors.New("the signature of Txn illegal")
)
