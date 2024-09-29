package MEVless

import (
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/pebble"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/types"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

const notifyBufferLen = 10

type MEVless struct {
	*tripod.Tripod
	cfg *Config

	commitmentsDB *pebble.DB

	notifyCh chan *OrderCommitment

	wsClients map[*websocket.Conn]bool
	wsLock    sync.Mutex
}

const Prefix = "MEVless_"

func NewMEVless(cfg *Config) (*MEVless, error) {
	db, err := pebble.Open(cfg.DbPath, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	tri := &MEVless{
		Tripod:        tripod.NewTripod(),
		cfg:           cfg,
		commitmentsDB: db,
		notifyCh:      make(chan *OrderCommitment, notifyBufferLen),
		wsClients:     make(map[*websocket.Conn]bool),
	}

	tri.SetWritings(tri.OrderTx)

	go tri.HandleSubscribe()
	return tri, nil
}

func (m *MEVless) CheckTxn(stxn *types.SignedTxn) error {
	// Just for print log
	hashStr := strings.TrimPrefix(stxn.GetParams(), Prefix)
	logrus.Printf("[OrderCommitment] Request Order Hash: %s\n]", hashStr)
	return nil
}

func (m *MEVless) OrderTx(ctx *context.WriteContext) error {
	return nil
}

func (m *MEVless) Pack(blockNum common.BlockNum, numLimit uint64) ([]*types.SignedTxn, error) {
	err := m.OrderCommitment(blockNum)
	if err != nil {
		return nil, err
	}
	return m.Pool.PackFor(numLimit, func(txn *types.SignedTxn) bool {
		return txn.ParamsIsJson()
	})
}

func (m *MEVless) PackFor(blockNum common.BlockNum, numLimit uint64, filter func(*types.SignedTxn) bool) ([]*types.SignedTxn, error) {
	err := m.OrderCommitment(blockNum)
	if err != nil {
		return nil, err
	}
	return m.Pool.PackFor(numLimit, func(txn *types.SignedTxn) bool {
		if txn.ParamsIsJson() {
			return filter(txn)
		}
		return false
	})
}

// wrCall.params = "MEVless_(TxnHash)"
func (m *MEVless) OrderCommitment(blockNum common.BlockNum) error {
	hashTxns, err := m.Pool.PackFor(m.cfg.PackNumber, func(txn *types.SignedTxn) bool {
		paramStr := txn.GetParams()
		return strings.HasPrefix(paramStr, Prefix)
	})
	if err != nil {
		return err
	}
	if len(hashTxns) == 0 {
		return nil
	}

	sequence := m.makeOrder(hashTxns)
	for i := 0; i < len(sequence); i++ {
		logrus.Printf("[OrderCommitment] makeOrder sequence: [%d] %v\n", i, sequence[i].Hex())
	}

	orderCommitment := &OrderCommitment{
		BlockNumber: blockNum,
		Sequences:   sequence,
	}

	m.notifyClient(orderCommitment)

	err = m.storeOrderCommitment(orderCommitment)
	if err != nil {
		return err
	}

	// TODO: sync the OrderCommitment to other P2P nodes

	// sleep for a while so that clients can send their tx-content onchain.
	time.Sleep(5000 * time.Millisecond)

	m.Pool.Reset(hashTxns)

	m.Pool.SortTxns(func(txs []*types.SignedTxn) []*types.SignedTxn {
		sorted := make([]*types.SignedTxn, 0)
		for i := 0; i < len(sequence); i++ {
			hash := sequence[i]
			for _, txn := range txs {
				if txn.TxnHash == hash {
					sorted = append(sorted, txn)
					txs = slices.DeleteFunc(txs, func(txn *types.SignedTxn) bool {
						return txn.TxnHash == hash
					})
					break
				}
			}
		}
		sorted = append(sorted, txs...)

		for num, seq := range sorted {
			logrus.Printf("[OrderCommitment] expected sequence: [%d] %v\n", num, seq.TxnHash.Hex())
		}

		return sorted
	})

	return nil
}

func (m *MEVless) makeOrder(hashTxns []*types.SignedTxn) map[int]common.Hash {
	order := make(map[int]common.Hash)
	sort.Slice(hashTxns, func(i, j int) bool {
		return hashTxns[i].GetTips() > hashTxns[j].GetTips()
	})
	for i, txn := range hashTxns {
		hashStr := strings.TrimPrefix(txn.GetParams(), Prefix)
		order[i] = common.HexToHash(hashStr)
	}
	return order
}

func (m *MEVless) VerifyBlock(block *types.Block) error {
	// TODO: verify tx order commitment from other miner node
	// TODO: for double-check, fetch txs from DA layers if the block does not have commitment txs.
	return nil
}

func (m *MEVless) Charge() uint64 {
	return m.cfg.Charge
}

type OrderCommitment struct {
	BlockNumber common.BlockNum     `json:"block_number"`
	Sequences   map[int]common.Hash `json:"sequences"`
}

type TxOrder struct {
	BlockNumber common.BlockNum `json:"block_number"`
	Sequence    int             `json:"sequence"`
}

func (m *MEVless) storeOrderCommitment(oc *OrderCommitment) error {
	batch := m.commitmentsDB.NewBatch()
	for seq, txnHash := range oc.Sequences {
		txOrder := &TxOrder{
			BlockNumber: oc.BlockNumber,
			Sequence:    seq,
		}
		byt, err := json.Marshal(txOrder)
		if err != nil {
			return err
		}
		err = batch.Set(txnHash.Bytes(), byt, pebble.NoSync)
		if err != nil {
			return err
		}
	}
	return batch.Commit(pebble.Sync)
}

func (m *MEVless) notifyClient(oc *OrderCommitment) {
	fmt.Printf("[NotifyClient] %#v\n", oc)
	select {
	case m.notifyCh <- oc:
	default:
		<-m.notifyCh
		m.notifyCh <- oc
	}
}
