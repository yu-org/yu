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
	ethStateConfig := setDefaultEthStateConfig()

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

//func emitReceipt(ctx *context.WriteContext, vmEvm *vm.EVM, txReq *TxRequest, contractAddr common.Address, leftOverGas uint64, err error) error {
//	evmReceipt := makeEvmReceipt(vmEvm, ctx.Txn, ctx.Block, contractAddr, leftOverGas, err)
//	receiptByt, err := json.Marshal(evmReceipt)
//	if err != nil {
//		return err
//	}
//	ctx.ExtraInterface = pd
//	return nil
//}

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

//func (s *Solidity) buyGas(state vm.StateDB, req *TxRequest) error {
//	gasFee := new(big.Int).Mul(req.GasPrice, new(big.Int).SetUint64(req.GasLimit))
//	gasFeeU256, _ := uint256.FromBig(gasFee)
//	if state.GetBalance(req.Origin).Cmp(gasFeeU256) < 0 {
//		return core.ErrInsufficientFunds
//	}
//	state.SubBalance(req.Origin, gasFeeU256, tracing.BalanceDecreaseGasBuy)
//	s.coinbaseReward.Add(gasFee.Uint64())
//	// return s.gasPool.SubGas(req.GasLimit)
//	return nil
//}

//func (s *Solidity) refundGas(state vm.StateDB, req *TxRequest, gasUsed uint64, refundQuotient uint64) {
//	refund := gasUsed / refundQuotient
//	if refund > state.GetRefund() {
//		refund = state.GetRefund()
//	}
//	remainGas := req.GasLimit - gasUsed + refund
//	refundFee := new(big.Int).Mul(req.GasPrice, new(big.Int).SetUint64(remainGas))
//	refundFeeU256, _ := uint256.FromBig(refundFee)
//	state.AddBalance(req.Origin, refundFeeU256, tracing.BalanceIncreaseGasReturn)
//	// s.gasPool.AddGas(remainGas)
//}

//func (s *Solidity) preCheck(req *TxRequest, stateDB vm.StateDB) error {
//	//// Make sure this transaction's nonce is correct.
//	//stNonce := stateDB.GetNonce(tx.Origin)
//	//fmt.Printf("From(%s) stateDB.Nonce = %d, request.Nonce = %d \n", tx.Origin.Hex(), stNonce, tx.Nonce)
//	//if msgNonce := tx.Nonce; stNonce < msgNonce {
//	//	return fmt.Errorf("%w: address %v, tx: %d state: %d", core.ErrNonceTooHigh,
//	//		tx.Origin.Hex(), msgNonce, stNonce)
//	//} else if stNonce > msgNonce {
//	//	return fmt.Errorf("%w: address %v, tx: %d state: %d", core.ErrNonceTooLow,
//	//		tx.Origin.Hex(), msgNonce, stNonce)
//	//} else if stNonce+1 < stNonce {
//	//	return fmt.Errorf("%w: address %v, nonce: %d", core.ErrNonceMax,
//	//		tx.Origin.Hex(), stNonce)
//	//}
//	//
//	//// Make sure the sender is an EOA
//	//codeHash := stateDB.GetCodeHash(tx.Origin)
//	//if codeHash != (common.Hash{}) && codeHash != types.EmptyCodeHash {
//	//	return fmt.Errorf("%w: address %v, codehash: %s", core.ErrSenderNoEOA,
//	//		tx.Origin.Hex(), codeHash)
//	//}
//	//
//	//return nil
//	//stNonce := stateDB.GetNonce(req.Origin)
//	//
//	//// fmt.Printf("address %s, tx.nonce: %d, state.nonce: %d \n", req.Origin.Hex(), req.Nonce, stNonce)
//	//if req.Nonce < stNonce {
//	//	return fmt.Errorf("%w: txHash: %s address %v, tx: %d state: %d", core.ErrNonceTooLow, req.Hash.String(),
//	//		req.Origin.Hex(), req.Nonce, stNonce)
//	//}
//	return s.buyGas(stateDB, req)
//}

//func (s *Solidity) executeContractCreation(ctx *context.WriteContext, txReq *TxRequest, stateDB *state.StateDB, origin, coinBase common.Address, vmenv *vm.EVM, sender vm.AccountRef, rules params.Rules) (uint64, error) {
//	stateDB.Prepare(rules, origin, coinBase, nil, vm.ActivePrecompiles(rules), nil)
//
//	code, address, leftOverGas, err := vmenv.Create(sender, txReq.Input, txReq.GasLimit, uint256.MustFromBig(txReq.Value))
//	if err != nil {
//		_ = emitReceipt(ctx, vmenv, txReq, code, address, leftOverGas, err)
//		return 0, err
//	}
//
//	return txReq.GasLimit - leftOverGas, emitReceipt(ctx, vmenv, txReq, code, address, leftOverGas, err)
//}

//func (s *Solidity) executeContractCall(ctx *context.WriteContext, txReq *TxRequest, ethState *state.StateDB, origin, coinBase common.Address, vmenv *vm.EVM, sender vm.AccountRef, rules params.Rules) (uint64, error) {
//	ethState.Prepare(rules, origin, coinBase, txReq.Address, vm.ActivePrecompiles(rules), nil)
//	ethState.SetNonce(txReq.Origin, ethState.GetNonce(txReq.Origin)+1, tracing.NonceChangeNewContract)
//
//	// logrus.Printf("before transfer: account %s balance %d \n", sender.Address(), ethState.GetBalance(sender.Address()))
//
//	code, leftOverGas, err := vmenv.Call(sender, *txReq.Address, txReq.Input, txReq.GasLimit, uint256.MustFromBig(txReq.Value))
//	// logrus.Printf("after transfer: account %s balance %d \n", sender.Address(), ethState.GetBalance(sender.Address()))
//	if err != nil {
//		// byt, _ := json.Marshal(txReq)
//		// logrus.Printf("[Execute Txn] SendTx Failed. err = %v. Request = %v", err, string(byt))
//		_ = emitReceipt(ctx, vmenv, txReq, code, common.Address{}, leftOverGas, err)
//		return 0, err
//	}
//
//	// logrus.Printf("[Execute Txn] SendTx success. Oringin code = %v, Hex Code = %v, Left Gas = %v", code, hex.EncodeToString(code), leftOverGas)
//	return txReq.GasLimit - leftOverGas, emitReceipt(ctx, vmenv, txReq, code, common.Address{}, leftOverGas, err)
//}
//
//func makeEvmReceipt(ctx *context.WriteContext, vmEvm *vm.EVM, code []byte, signedTx *yu_types.SignedTxn, block *yu_types.Block, address common.Address, leftOverGas uint64, err error) *types.Receipt {
//	wrCallParams := signedTx.Raw.WrCall.Params
//	txReq := &TxRequest{}
//	_ = json.Unmarshal([]byte(wrCallParams), txReq)
//
//	txArgs := &TempTransactionArgs{}
//	_ = json.Unmarshal(txReq.OriginArgs, txArgs)
//	originTx := txArgs.ToTransaction(txReq.V, txReq.R, txReq.S)
//
//	stateDb := vmEvm.StateDB.(*pending_state.PendingStateWrapper).GetStateDB()
//	usedGas := originTx.Gas() - leftOverGas
//
//	blockNumber := big.NewInt(int64(block.Height))
//	txHash := common.Hash(signedTx.TxnHash)
//	effectiveGasPrice := big.NewInt(1000000000) // 1 GWei
//
//	status := types.ReceiptStatusFailed
//	if err == nil {
//		status = types.ReceiptStatusSuccessful
//	}
//	var root []byte
//	//stateDB := vmEvm.StateDB.(*pending_state.PendingState)
//	//if vmEvm.ChainConfig().IsByzantium(blockNumber) {
//	//	stateDB.Finalise(true)
//	//} else {
//	//	root = stateDB.IntermediateRoot(vmEvm.ChainConfig().IsEIP158(blockNumber)).Bytes()
//	//}
//
//	// TODO: 1. root is nil; 2. CumulativeGasUsed not; 3. logBloom is empty
//
//	receipt := &types.Receipt{
//		Type:              originTx.Type(),
//		Status:            status,
//		PostState:         root,
//		CumulativeGasUsed: leftOverGas,
//		TxHash:            txHash,
//		ContractAddress:   address,
//		GasUsed:           usedGas,
//		EffectiveGasPrice: effectiveGasPrice,
//	}
//
//	if originTx.Type() == types.BlobTxType {
//		receipt.BlobGasUsed = uint64(len(originTx.BlobHashes()) * params.BlobTxBlobGasPerBlob)
//		receipt.BlobGasPrice = vmEvm.Context.BlobBaseFee
//	}
//
//	receipt.Logs = stateDb.GetLogs(txHash, blockNumber.Uint64(), common.Hash(block.Hash))
//	receipt.Bloom = types.CreateBloom(types.Receipts{})
//	receipt.BlockHash = common.Hash(block.Hash)
//	receipt.BlockNumber = blockNumber
//	receipt.TransactionIndex = uint(ctx.TxnIndex)
//
//	// spew.Dump("[Receipt] log = %v", stateDB.Logs())
//	// logrus.Printf("[Receipt] log is nil = %v", receipt.Logs == nil)
//	if receipt.Logs == nil {
//		receipt.Logs = []*types.Log{}
//	}
//
//	for idx, txn := range block.Txns {
//		if common.Hash(txn.TxnHash) == txHash {
//			receipt.TransactionIndex = uint(idx)
//		}
//	}
//	// logrus.Printf("[Receipt] statedb txIndex = %v, actual txIndex = %v", ctx.TxnIndex, receipt.TransactionIndex)
//
//	return receipt
//}

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

//func checkGetReceipt() (checkResult bool) {
//	limiter := utils.GetReceiptRateLimiter
//	if config.GetGlobalConfig().RateLimitConfig.GetReceipt < 1 || limiter == nil {
//		return true
//	}
//	if !limiter.Allow() {
//		return false
//	}
//	if err := limiter.Wait(context2.Background()); err != nil {
//		return false
//	}
//	return true
//}

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

//func emitReceipt(ctx *context.WriteContext, vmEmv *vm.EVM, txReq *TxRequest, code []byte, contractAddr common.Address, leftOverGas uint64, err error) error {
//	evmReceipt := makeEvmReceipt(ctx, vmEmv, code, ctx.Txn, ctx.Block, contractAddr, leftOverGas, err)
//	var buf bytes.Buffer
//	encodeErr := json.NewEncoder(&buf).Encode(evmReceipt)
//	if encodeErr != nil {
//		logrus.Errorf("Receipt marshal err: %v. Tx: %s", encodeErr, txReq.Hash())
//		return encodeErr
//	}
//	ctx.EmitExtra(buf.Bytes())
//	return nil
//}

var ErrNotFoundReceipt = errors.New("receipt not found")

func (s *Solidity) applyEVM(evm *vm.EVM, gp *core.GasPool, db *state.StateDB, block *yu_types.Block, tx *types.Transaction, usedGas *uint64) (*types.Receipt, error) {
	msg, err := core.TransactionToMessage(tx, types.MakeSigner(evm.ChainConfig(), block.Height.ToBigInt(), block.Timestamp), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return core.ApplyTransactionWithEVM(msg, gp, db, block.Height.ToBigInt(), common.Hash(block.Hash), block.Timestamp, tx, usedGas, evm)
}

// endregion ---- Tripod Api ----
