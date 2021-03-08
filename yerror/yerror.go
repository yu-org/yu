package yerror

import (
	"github.com/pkg/errors"
	. "yu/common"
)

var TypeErr = errors.New("the type of params error")

var NoKeyType = errors.New("no key type")
var NoKvdbType = errors.New("no kvdb type")
var NoQueueType = errors.New("no queue type")

var (
	PoolOverflow    error = errors.New("pool size is full")
	TxnSignatureErr error = errors.New("the signature of Txn illegal")
	TxnTooLarge     error = errors.New("the size of txn is too large")
	TxnDuplicate    error = errors.New("txn is duplicate")
)

type ErrTripodNotFound struct {
	TripodName string
}

func TripodNotFound(name string) ErrTripodNotFound {
	return ErrTripodNotFound{TripodName: name}
}

func (t ErrTripodNotFound) Error() string {
	return errors.Errorf("Tripod (%s) NOT Found", t.TripodName).Error()
}

type ErrExecNotFound struct {
	ExecName string
}

func ExecNotFound(name string) ErrExecNotFound {
	return ErrExecNotFound{ExecName: name}
}

func (e ErrExecNotFound) Error() string {
	return errors.Errorf("Execution(%s) NOT Found", e.ExecName).Error()
}

type ErrQryNotFound struct {
	QryName string
}

func QryNotFound(name string) ErrQryNotFound {
	return ErrQryNotFound{QryName: name}
}

func (q ErrQryNotFound) Error() string {
	return errors.Errorf("Query(%s) NOT Found", q.QryName).Error()
}

type ErrNodeKeeperDead struct {
	IP string
}

func NodeKeeperDead(ip string) ErrNodeKeeperDead {
	return ErrNodeKeeperDead{IP: ip}
}

func (n ErrNodeKeeperDead) Error() string {
	return errors.Errorf("NodeKeeper(%s) is dead", n.IP).Error()
}

type ErrWorkerDead struct {
	Name string
}

func WorkerDead(name string) ErrWorkerDead {
	return ErrWorkerDead{Name: name}
}

func (w ErrWorkerDead) Error() string {
	return errors.Errorf("Worker(%s) is dead", w.Name).Error()
}

type ErrWaitTxnsTimeout struct {
	TxnsHash []Hash
}

func WaitTxnsTimeout(hashesMap map[Hash]bool) ErrWaitTxnsTimeout {
	hashes := make([]Hash, 0)
	for hash, _ := range hashesMap {
		hashes = append(hashes, hash)
	}
	return ErrWaitTxnsTimeout{TxnsHash: hashes}
}

func (wt ErrWaitTxnsTimeout) Error() string {
	return errors.Errorf("waiting txns-hashes timeout: %v", wt.TxnsHash).Error()
}
