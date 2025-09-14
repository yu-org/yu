package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	yu_common "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/tripod"
	yu_types "github.com/yu-org/yu/core/types"

	"github.com/yu-org/yu/apps/eth/config"
	"github.com/yu-org/yu/apps/eth/metrics"
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
)

type Solidity struct {
	sync.Mutex

	*tripod.Tripod
	ethState    *EthState
	cfg         *config.GethConfig
	stateConfig *config.Config

	// gasPool        *core.GasPool
	coinbaseReward atomic.Uint64
}

//func (s *Solidity) StateDB() *state.StateDB {
//	s.Lock()
//	defer s.Unlock()
//	return s.ethState.StateDB()
//}

func copyEvmFromRequest(cfg *config.GethConfig, req *TxRequest) *vm.EVM {
	//txContext := vm.TxContext{
	//	Origin:     req.Origin,
	//	GasPrice:   req.GasPrice,
	//	BlobHashes: cfg.BlobHashes,
	//	BlobFeeCap: cfg.BlobFeeCap,
	//}
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     cfg.GetHashFn,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    req.Gas(),
		BaseFee:     cfg.BaseFee,
		BlobBaseFee: cfg.BlobBaseFee,
		Random:      cfg.Random,
	}

	return vm.NewEVM(blockContext, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}

func newEVM(cfg *config.GethConfig) *vm.EVM {
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     cfg.GetHashFn,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		Difficulty:  cfg.Difficulty,
		GasLimit:    cfg.GasLimit,
		BaseFee:     cfg.BaseFee,
		BlobBaseFee: cfg.BlobBaseFee,
		Random:      cfg.Random,
	}

	return vm.NewEVM(blockContext, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}

func (s *Solidity) InitChain(genesisBlock *yu_types.Block) {
	cfg := s.stateConfig
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

	ethState, err := NewEthState(cfg, lastStateRoot)
	if err != nil {
		logrus.Fatal("init NewEthState failed: ", err)
	}
	s.ethState = ethState
	s.cfg.State = ethState.stateDB

	_, _, _, err = SetupGenesisBlock(ethState.ethDB, ethState.trieDB, genesis)
	if err != nil {
		logrus.Fatal("SetupGenesisBlock failed: ", err)
	}

	// s.cfg.ChainConfig = chainConfig

	// commit genesis state
	genesisStateRoot, err := s.ethState.GenesisCommit()
	if err != nil {
		logrus.Fatal("genesis state commit failed: ", err)
	}

	genesisBlock.StateRoot = yu_common.Hash(genesisStateRoot)
}

func NewSolidity(gethConfig *config.GethConfig) *Solidity {
	ethStateConfig := config.SetDefaultEthStateConfig()

	solidity := &Solidity{
		Tripod:      tripod.NewTripod(),
		cfg:         gethConfig,
		stateConfig: ethStateConfig,
		// network:       utils.Network(cfg.Network),
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
	s.Lock()
	start := time.Now()
	defer func() {
		s.Unlock()
		metrics.SolidityHist.WithLabelValues(startBlockLbl).Observe(float64(time.Since(start).Microseconds()))
	}()
	s.cfg.BlockNumber = big.NewInt(int64(block.Height))
	// s.gasPool = new(core.GasPool).AddGas(block.LeiLimit)
	s.cfg.GasLimit = block.LeiLimit
	s.cfg.Time = block.Timestamp
	s.cfg.Difficulty = big.NewInt(int64(block.Difficulty))
}

func (s *Solidity) EndBlock(block *yu_types.Block) {
	// nothing
}

func (s *Solidity) FinalizeBlock(block *yu_types.Block) {
	// nothing
}

func (s *Solidity) PreHandleTxn(txn *yu_types.SignedTxn) error {
	var txReq TxRequest
	param := txn.GetParams()
	err := json.Unmarshal([]byte(param), &txReq)
	if err != nil {
		return err
	}
	yuHash, err := ConvertHashToYuHash(txReq.Hash())
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
	s.Lock()
	start := time.Now()
	defer func() {
		s.Unlock()
		metrics.SolidityHist.WithLabelValues(executeTxnLbl).Observe(float64(time.Since(start).Microseconds()))
		if err == nil {
			metrics.SolidityCounter.WithLabelValues(executeTxnLbl, statusSuccess).Inc()
		} else {
			metrics.SolidityCounter.WithLabelValues(executeTxnLbl, statusErr).Inc()
		}
	}()
	txReq := new(TxRequest)
	// coinbase := common.BytesToAddress(s.cfg.Coinbase.Bytes())

	_ = ctx.BindJson(txReq)
	evm := copyEvmFromRequest(s.cfg, txReq)
	gasPool := new(core.GasPool).AddGas(ctx.Block.LeiLimit)

	s.ethState.stateDB.SetTxContext(txReq.Hash(), ctx.TxnIndex)
	rcpt, err := s.applyEVM(evm, gasPool, s.ethState.stateDB, ctx.Block, txReq.Transaction, nil)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	encodeErr := json.NewEncoder(&buf).Encode(rcpt)
	if encodeErr != nil {
		logrus.Errorf("Receipt marshal err: %v. Tx: %s", encodeErr, txReq.Hash())
		return encodeErr
	}
	ctx.EmitExtra(buf.Bytes())
	return
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
func (s *Solidity) Call(ctx *context.ReadContext) {
	metrics.SolidityCounter.WithLabelValues(callTxnLbl, statusSuccess).Inc()
	s.Lock()
	start := time.Now()
	defer func() {
		s.Unlock()
		metrics.SolidityHist.WithLabelValues(callTxnLbl).Observe(float64(time.Since(start).Microseconds()))
	}()

	callReq := new(CallRequest)
	err := ctx.BindJson(callReq)
	if err != nil {
		ctx.Json(http.StatusBadRequest, &CallResponse{Err: err})
		return
	}
	address := callReq.Address
	input := callReq.Input
	origin := callReq.Origin
	gasLimit := callReq.GasLimit
	gasPrice := callReq.GasPrice
	value := callReq.Value

	cfg := s.cfg
	cfg.Origin = origin
	cfg.GasLimit = gasLimit
	cfg.GasPrice = gasPrice
	cfg.Value = value
	ethState := s.ethState

	var (
		vmenv = newEVM(cfg)
		rules = cfg.ChainConfig.Rules(vmenv.Context.BlockNumber, vmenv.Context.Random != nil, vmenv.Context.Time)
	)
	vmenv.StateDB = s.ethState.StateDB().Copy()
	if cfg.EVMConfig.Tracer != nil && cfg.EVMConfig.Tracer.OnTxStart != nil {
		cfg.EVMConfig.Tracer.OnTxStart(vmenv.GetVMContext(), types.NewTx(&types.LegacyTx{To: &address, Data: input, Value: value, Gas: gasLimit}), origin)
	}
	// Execute the preparatory steps for state transition which includes:
	// - prepare accessList(post-berlin)
	// - reset transient storage(eip 1153)
	ethState.Prepare(rules, origin, cfg.Coinbase, &address, vm.ActivePrecompiles(rules), nil)

	// Call the code with the given configuration.
	ret, leftOverGas, err := vmenv.Call(
		origin,
		address,
		input,
		gasLimit,
		uint256.MustFromBig(value),
	)

	logrus.Debugf("[Call] Request from = %v, to = %v, gasLimit = %v, value = %v, input = %v", origin.Hex(), address.Hex(), gasLimit, value.Uint64(), hex.EncodeToString(input))
	logrus.Debugf("[Call] Response: Origin Code = %v, Hex Code = %v, String Code = %v, LeftOverGas = %v", ret, hex.EncodeToString(ret), new(big.Int).SetBytes(ret).String(), leftOverGas)

	if err != nil {
		ctx.Json(http.StatusInternalServerError, &CallResponse{Err: err})
		return
	}
	result := CallResponse{Ret: ret, LeftOverGas: leftOverGas}
	ctx.JsonOk(&result)
}

func (s *Solidity) Commit(block *yu_types.Block) {
	metrics.SolidityCounter.WithLabelValues(commitLbl, statusSuccess).Inc()
	s.Lock()
	start := time.Now()
	defer func() {
		s.Unlock()
		metrics.SolidityHist.WithLabelValues(commitLbl).Observe(float64(time.Since(start).Microseconds()))
	}()

	// reward coinbase
	s.ethState.AddBalance(s.cfg.Coinbase, uint256.NewInt(s.coinbaseReward.Load()), tracing.BalanceIncreaseRewardTransactionFee)
	s.coinbaseReward.Store(0)

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
	s.Lock()
	defer s.Unlock()
	sdb, err := s.ethState.StateAt(root)
	if err != nil {
		return nil, err
	}
	return sdb.Copy(), nil
}

func (s *Solidity) GetEthDB() ethdb.Database {
	return s.ethState.ethDB
}

type ReceiptRequest struct {
	Hash common.Hash `json:"hash"`
}

type ReceiptResponse struct {
	Receipt *types.Receipt `json:"receipt"`
	Err     error          `json:"err"`
}

type ReceiptsRequest struct {
	Hashes []common.Hash `json:"hashes"`
}

type ReceiptsResponse struct {
	Receipts []*types.Receipt `json:"receipts"`
	Err      error            `json:"err"`
}

func (s *Solidity) GetEthReceipt(hash common.Hash) (*types.Receipt, error) {
	yuHash, err := ConvertHashToYuHash(hash)
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
		return nil, ErrNotFoundReceipt
	}

	// logrus.Printf("yuReceipt.Extra(%s): %s", yuHash.String(), string(yuReceipt.Extra))

	receipt := new(types.Receipt)
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
	if !ValidateTxHash(rq.Hash.Hex()) {
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
		if !ValidateTxHash(hash.Hex()) {
			metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusErr).Inc()
			continue
		}
		yuHash, err := ConvertHashToYuHash(hash)
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
	want := make([]*types.Receipt, 0)
	for _, yuRecipt := range yuReceipts {
		receipt := new(types.Receipt)
		if yuRecipt.Extra != nil {
			json.NewDecoder(bytes.NewBuffer(yuRecipt.Extra)).Decode(receipt)
		}
		want = append(want, receipt)
	}
	metrics.SolidityCounter.WithLabelValues(getReceiptsLbl, statusSuccess).Inc()
	ctx.JsonOk(&ReceiptsResponse{Receipts: want})
}

var ErrNotFoundReceipt = errors.New("receipt not found")

func (s *Solidity) applyEVM(evm *vm.EVM, gp *core.GasPool, db *state.StateDB, block *yu_types.Block, tx *types.Transaction, usedGas *uint64) (*types.Receipt, error) {
	msg, err := core.TransactionToMessage(tx, types.MakeSigner(evm.ChainConfig(), block.Height.ToBigInt(), block.Timestamp), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return core.ApplyTransactionWithEVM(msg, gp, db, block.Height.ToBigInt(), common.Hash(block.Hash), block.Timestamp, tx, usedGas, evm)
}

// endregion ---- Tripod Api ----
