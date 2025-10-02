package evm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/yu-org/yu/apps/eth/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/sirupsen/logrus"
	yu_common "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/tripod"
	yu_types "github.com/yu-org/yu/core/types"

	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/eth/metrics"
	"github.com/yu-org/yu/apps/eth/types"
)

var (
	startBlockLbl  = "start"
	finaliseLbl    = "finalise"
	setStateDBLbl  = "set"
	executeTxnLbl  = "execute"
	callTxnLbl     = "call"
	commitLbl      = "commit"
	getReceiptsLbl = "gets"
	getReceiptLbl  = "get"

	statusSuccess = "success"
	statusErr     = "err"
	statusExceed  = "exceed"
	debugAddr     = common.HexToAddress("0x7Bd36074b61Cfe75a53e1B9DF7678C96E6463b02")
)

type Solidity struct {
	*tripod.Tripod
	ethState *EthState
	cfg      *config.GethConfig

	gasPool *core.GasPool
}

func (s *Solidity) InitChain(genesisBlock *yu_types.Block) {
	var genesis *Genesis
	if s.cfg.IsReddioMainnet {
		genesis = DefaultGenesisBlock()
	} else {
		genesis = DefaultSepoliaGenesisBlock()
	}

	var lastStateRoot common.Hash
	block, err := s.GetCurrentBlock()
	if err != nil && err != yerror.ErrBlockNotFound {
		logrus.Fatal("get current block failed: ", err)
	}
	if block != nil {
		lastStateRoot = common.Hash(block.StateRoot)
	}

	ethState, err := NewEthState(lastStateRoot, s.cfg)
	if err != nil {
		logrus.Fatal("init NewEthState failed: ", err)
	}
	s.ethState = ethState

	// commit genesis state
	genesisStateRoot, err := s.ethState.GenesisCommit(genesis)
	if err != nil {
		logrus.Fatal("genesis state commit failed: ", err)
	}

	genesisBlock.StateRoot = yu_common.Hash(genesisStateRoot)
}

func NewSolidity(gethConfig *config.GethConfig) *Solidity {

	solidity := &Solidity{
		Tripod: tripod.NewTripod(),
		cfg:    gethConfig,
	}
	solidity.SetWritings(solidity.ExecuteTxn)
	solidity.SetReadings(
		solidity.Call, solidity.GetReceipt, solidity.GetReceipts,
		// solidity.GetClass, solidity.GetClassAt,
		// 	solidity.GetClassHashAt, solidity.GetNonce, solidity.GetStorage,
		// 	solidity.GetTransaction, solidity.GetTransactionStatus,
		// 	solidity.SimulateTransactions,
		// 	solidity.GetBlockWithTxs, solidity.GetBlockWithTxHashes,
	)

	return solidity
}

// region ---- Tripod Api ----

func (s *Solidity) StartBlock(block *yu_types.Block) {
	metrics.SolidityCounter.WithLabelValues(startBlockLbl, statusSuccess).Inc()
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(startBlockLbl).Observe(float64(time.Since(start).Microseconds()))
	}()
	s.cfg.BlockNumber = big.NewInt(int64(block.Height))
	s.cfg.GasLimit = block.LeiLimit
	s.cfg.Time = block.Timestamp
	s.gasPool = new(core.GasPool).AddGas(block.LeiLimit)
	s.cfg.Difficulty = big.NewInt(int64(block.Difficulty))
	err := s.ethState.StartState()
	if err != nil {
		logrus.Errorf("start state at Block(%d) failed: %v", block.Height, err)
	}
}

func (s *Solidity) EndBlock(block *yu_types.Block) {
	// nothing
}

func (s *Solidity) FinalizeBlock(block *yu_types.Block) {
	// nothing
}

func (s *Solidity) PreHandleTxn(txn *yu_types.SignedTxn) error {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("Recover, txn hash: %s, err: %v", txn.TxnHash.String(), err)
		}
	}()
	param := txn.GetParams()
	txReq, err := DecodeTxReq([]byte(param))
	if err != nil {
		return err
	}
	yuHash, err := utils.ConvertHashToYuHash(txReq.ToEthTx().Hash())
	if err != nil {
		return err
	}
	txn.TxnHash = yuHash
	return nil
}

func (s *Solidity) CheckTxn(txn *yu_types.SignedTxn) error {
	req := new(TxRequest)
	err := txn.BindJson(req)
	if err != nil {
		return err
	}

	if req.IsInternalCall {
		// TODO: use txn.Pubkey and txn.Signature to verify the tx
	}
	return nil
}

// ExecuteTxn executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Execute sets up an in-memory, temporary, environment for the execution of
// the given code. It makes sure that it's restored to its original state afterwards.
func (s *Solidity) ExecuteTxn(ctx *context.WriteContext) (err error) {
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(executeTxnLbl).Observe(float64(time.Since(start).Microseconds()))
		if err == nil {
			metrics.SolidityCounter.WithLabelValues(executeTxnLbl, statusSuccess).Inc()
		} else {
			metrics.SolidityCounter.WithLabelValues(executeTxnLbl, statusErr).Inc()
		}
	}()

	logrus.Infof("ExecuteTxn, debugAddr: %s amount: %d", debugAddr.Hex(), s.ethState.stateDB.GetBalance(debugAddr))

	txReq, err := DecodeTxReq(ctx.GetRequestBytes())
	if err != nil {
		return err
	}
	rcpt, err := s.ethState.ApplyTx(ctx.Block, txReq.ToEthTx(), ctx.TxnIndex, s.gasPool, new(uint64))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encodeErr := json.NewEncoder(&buf).Encode(rcpt)
	if encodeErr != nil {
		logrus.Errorf("Receipt marshal err: %v. Tx: %s", encodeErr, txReq.ToEthTx().Hash())
		return encodeErr
	}
	ctx.EmitExtra(buf.Bytes())
	return
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
func (s *Solidity) Call(ctx *context.ReadContext) {
	metrics.SolidityCounter.WithLabelValues(callTxnLbl, statusSuccess).Inc()
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(callTxnLbl).Observe(float64(time.Since(start).Microseconds()))
	}()

	callReq := new(types.CallRequest)
	err := ctx.BindJson(callReq)
	if err != nil {
		ctx.Json(http.StatusBadRequest, &types.CallResponse{Err: err})
		return
	}

	msg := callReq.TxArgs.ToMessage(s.cfg.BaseFee, false, false)
	res, err := s.ethState.ApplyTxForReader(msg)
	if err != nil {
		ctx.Json(http.StatusInternalServerError, &types.CallResponse{Err: err})
		return
	}

	result := types.CallResponse{Ret: res.ReturnData}
	ctx.JsonOk(&result)
}

func (s *Solidity) Commit(block *yu_types.Block) {
	metrics.SolidityCounter.WithLabelValues(commitLbl, statusSuccess).Inc()
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(commitLbl).Observe(float64(time.Since(start).Microseconds()))
	}()

	blockNumber := uint64(block.Height)
	stateRoot, err := s.ethState.Commit(blockNumber)
	if err != nil {
		logrus.Errorf("Solidity commit failed on Block(%d), error: %v", blockNumber, err)
		return
	}
	block.StateRoot = yu_common.Hash(stateRoot)
	// s.gasPool.SetGas(0)
}

func (s *Solidity) StateAt(root common.Hash) (*state.StateDB, error) {
	sdb, err := s.ethState.StateAt(root)
	if err != nil {
		return nil, err
	}
	return sdb.Copy(), nil
}

func (s *Solidity) GetEthDB() ethdb.Database {
	return s.ethState.db
}

type ReceiptRequest struct {
	Hash common.Hash `json:"hash"`
}

type ReceiptResponse struct {
	Receipt *ethtypes.Receipt `json:"receipt"`
	Err     error             `json:"err"`
}

type ReceiptsRequest struct {
	Hashes []common.Hash `json:"hashes"`
}

type ReceiptsResponse struct {
	Receipts []*ethtypes.Receipt `json:"receipts"`
	Err      error               `json:"err"`
}

func (s *Solidity) GetEthReceipt(hash common.Hash) (*ethtypes.Receipt, error) {
	yuHash, err := utils.ConvertHashToYuHash(hash)
	if err != nil {
		return nil, err
	}
	yuReceipt, err := s.TxDB.GetReceipt(yuHash)
	if err != nil {
		logrus.Debugf("getReceipt() TxDB.GetReceipt, txHash(%s) error: %v", yuHash.String(), err)
		return nil, err
	}

	// logrus.Printf("yuReceipt body is %s", yuReceipt.String())

	if yuReceipt == nil {
		logrus.Warnf("getReceipt() TxDB.GetReceipt, txHash(%s) not found，hash(%s)", yuHash.String(), hash.String())
		return nil, utils.ErrNotFoundReceipt
	}

	// logrus.Printf("yuReceipt.Extra(%s): %s", yuHash.String(), string(yuReceipt.Extra))

	receipt := new(ethtypes.Receipt)
	if yuReceipt.Extra == nil {
		return receipt, nil
	}
	err = json.NewDecoder(bytes.NewBuffer(yuReceipt.Extra)).Decode(receipt)
	if err != nil {
		logrus.Errorf("json.Unmarshal yuReceipt.Extra(%s) failed: %v", yuHash.String(), err)
	}
	return receipt, err
}

func (s *Solidity) GetReceipt(ctx *context.ReadContext) {
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(getReceiptLbl).Observe(float64(time.Since(start).Microseconds()))
	}()
	var rq ReceiptRequest
	err := ctx.BindJson(&rq)
	if err != nil {
		metrics.SolidityCounter.WithLabelValues(getReceiptLbl, statusErr).Inc()
		ctx.Json(http.StatusBadRequest, &ReceiptResponse{Err: fmt.Errorf("Solidity.GetReceipt parse json error:%v", err)})
		return
	}
	if !utils.ValidateTxHash(rq.Hash.Hex()) {
		metrics.SolidityCounter.WithLabelValues(getReceiptLbl, statusErr).Inc()
		ctx.Json(http.StatusBadRequest, &ReceiptResponse{Err: fmt.Errorf("Solidity.GetReceipt ValidateTxHash json error:%v", err)})
		return
	}
	receipt, err := s.GetEthReceipt(rq.Hash)
	if err != nil {
		metrics.SolidityCounter.WithLabelValues(getReceiptLbl, statusErr).Inc()
		ctx.Json(http.StatusInternalServerError, &ReceiptResponse{Err: err})
		return
	}
	metrics.SolidityCounter.WithLabelValues(getReceiptLbl, statusSuccess).Inc()
	ctx.JsonOk(&ReceiptResponse{Receipt: receipt})
}

func (s *Solidity) GetReceipts(ctx *context.ReadContext) {
	start := time.Now()
	defer func() {
		metrics.SolidityHist.WithLabelValues(getReceiptsLbl).Observe(float64(time.Since(start).Microseconds()))
	}()
	var rq ReceiptsRequest
	err := ctx.BindJson(&rq)
	if err != nil {
		metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusErr).Inc()
		ctx.Json(http.StatusBadRequest, &ReceiptsResponse{Err: fmt.Errorf("Solidity.GetReceipts parse json error:%v", err)})
		return
	}
	yuHashList := make([]yu_common.Hash, 0)
	for _, hash := range rq.Hashes {
		if !utils.ValidateTxHash(hash.Hex()) {
			metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusErr).Inc()
			continue
		}
		yuHash, err := utils.ConvertHashToYuHash(hash)
		if err != nil {
			metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusErr).Inc()
			ctx.Json(http.StatusBadRequest, &ReceiptsResponse{Err: fmt.Errorf("Solidity.GetReceipts parse json error:%v", err)})
			return
		}
		yuHashList = append(yuHashList, yuHash)
	}
	yuReceipts, err := s.TxDB.GetReceipts(yuHashList)
	if err != nil {
		metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusErr).Inc()
		ctx.Json(http.StatusBadRequest, &ReceiptsResponse{Err: fmt.Errorf("Solidity.GetReceipts parse json error:%v", err)})
		return
	}
	want := make([]*ethtypes.Receipt, 0)
	for _, yuRecipt := range yuReceipts {
		receipt := new(ethtypes.Receipt)
		if yuRecipt.Extra != nil {
			json.NewDecoder(bytes.NewBuffer(yuRecipt.Extra)).Decode(receipt)
		}
		want = append(want, receipt)
	}

	metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusSuccess).Inc()
	ctx.JsonOk(&ReceiptsResponse{Receipts: want})
}

func (s *Solidity) applyEVM(evm *vm.EVM, gp *core.GasPool, db *state.StateDB, block *yu_types.Block, tx *ethtypes.Transaction, usedGas *uint64) (*ethtypes.Receipt, error) {
	msg, err := core.TransactionToMessage(tx, ethtypes.MakeSigner(evm.ChainConfig(), block.Height.ToBigInt(), block.Timestamp), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return core.ApplyTransactionWithEVM(msg, gp, db, block.Height.ToBigInt(), common.Hash(block.Hash), block.Timestamp, tx, usedGas, evm)
}

// endregion ---- Tripod Api ----
