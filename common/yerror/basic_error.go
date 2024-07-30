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
var NoSqlDbType = errors.New("no sqlDB type")

var (
	PoolOverflow  error = errors.New("pool size is full")
	TxnTimeoutErr error = errors.New("Txn time out")
	TxnTooLarge   error = errors.New("the size of txn is too large")
	TxnDuplicated error = errors.New("Transaction duplicated")
)

var ErrBlockNotFound error = errors.New("block not found")

var OutOfLei = errors.New("Lei out")

type ErrTxnNotFound struct {
	txHash Hash
}

func TxnNotFound(txHash Hash) ErrTxnNotFound {
	return ErrTxnNotFound{txHash: txHash}
}

func (e ErrTxnNotFound) Error() string {
	return errors.Errorf("txn (%s) not found", e.txHash).Error()
}

type ErrBlockSignatureIllegal struct {
	blockHash Hash
}

func BlockSignatureIllegal(blockHash Hash) ErrBlockSignatureIllegal {
	return ErrBlockSignatureIllegal{blockHash: blockHash}
}

func (e ErrBlockSignatureIllegal) Error() string {
	return errors.Errorf("the signature of block(%s) is illegal", e.blockHash).Error()
}

type ErrTxnSignatureIllegal struct {
	err error
}

func TxnSignatureIllegal(err error) ErrTxnSignatureIllegal {
	return ErrTxnSignatureIllegal{err: err}
}

func (ErrTxnSignatureIllegal) Error() string {
	return errors.Errorf("txn signature illegal").Error()
}

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
	return errors.Errorf("Tripod(%s) NOT Found", t.TripodName).Error()
}

type ErrWritingNotFound struct {
	WritingName string
}

func WritingNotFound(name string) ErrWritingNotFound {
	return ErrWritingNotFound{WritingName: name}
}

func (e ErrWritingNotFound) Error() string {
	return errors.Errorf("Writing(%s) NOT Found", e.WritingName).Error()
}

type ErrReadingNotFound struct {
	ReadingName string
}

func ReadingNotFound(name string) ErrReadingNotFound {
	return ErrReadingNotFound{ReadingName: name}
}

func (q ErrReadingNotFound) Error() string {
	return errors.Errorf("Reading(%s) NOT Found", q.ReadingName).Error()
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
