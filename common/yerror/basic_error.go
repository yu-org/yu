package yerror

import (
	"github.com/pkg/errors"
	. "github.com/yu-org/yu/common"
)

var TypeErr = errors.New("the type of params error")
var IntegerOverflow = errors.New("integer overflow")

var NoP2PTopic = errors.New("no p2p topic")

var NoRunMode = errors.New("no run mode")
var NoKeyType = errors.New("no key type")
var NoConvergeType = errors.New("no converge type")

var GenesisBlockIllegal = errors.New("genesis block is illegal")

var NoKvdbType = errors.New("no kvdb type")
var NoQueueType = errors.New("no queue type")
var NoSqlDbType = errors.New("no sqlDB type")

var (
	PoolOverflow    error = errors.New("pool size is full")
	TxnSignatureErr error = errors.New("the signature of Txn illegal")
	TxnTimeoutErr   error = errors.New("Txn time out")
	TxnTooLarge     error = errors.New("the size of txn is too large")
)

var OutOfLei = errors.New("Lei out")

type ErrBlockIllegal struct {
	BlockHash string
}

func BlockIllegal(blockHash Hash) ErrBlockIllegal {
	return ErrBlockIllegal{BlockHash: blockHash.String()}
}

func (b ErrBlockIllegal) Error() string {
	return errors.Errorf("block(%s) illegal", b.BlockHash).Error()
}

type ErrNoTxnInP2P struct {
	TxnHash string
}

func NoTxnInP2P(txnHash Hash) ErrNoTxnInP2P {
	return ErrNoTxnInP2P{TxnHash: txnHash.String()}
}

func (t ErrNoTxnInP2P) Error() string {
	return errors.Errorf("no txn(%s) in P2P network", t.TxnHash).Error()
}

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
	return errors.Errorf("Read(%s) NOT Found", q.QryName).Error()
}

//type ErrOutOfEnergy struct {
//	txnsHashes []string
//}
//
//func OutOfLei(txnsHashes []Hash) ErrOutOfEnergy {
//	hashes := make([]string, 0)
//	for _, txnHash := range txnsHashes {
//		hashes = append(hashes, txnHash.String())
//	}
//	return ErrOutOfEnergy{txnsHashes: hashes}
//}
//
//func (oe ErrOutOfEnergy) Error() string {
//	return errors.Errorf("Energy out! txns(%v) ")
//}

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
