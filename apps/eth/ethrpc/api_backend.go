package ethrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yucommon "github.com/yu-org/yu/common"
	yucontext "github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/kernel"
	"github.com/yu-org/yu/core/protocol"
	yutypes "github.com/yu-org/yu/core/types"

	"github.com/reddio-com/reddio/evm"
)

type EthAPIBackend struct {
	allowUnprotectedTxs bool
	ethChainCfg         *params.ChainConfig
	chain               *kernel.Kernel
	gasPriceCache       *EthGasPrice
}

const (
	MaxRetries      = 3
	RetryIntervalMs = 500 * time.Millisecond
)

func (e *EthAPIBackend) SyncProgress() ethereum.SyncProgress {
	// TODO implement me
	panic("implement me")
}

//func (e *EthAPIBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
//	//TODO implement me
//	panic("implement me")
//}

// BlobBaseFee Move to ethrpc/gasprice.go
// func (e *EthAPIBackend) FeeHistory(ctx context.Context, blockCount uint64, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (*big.Int, [][]*big.Int, []*big.Int, []float64, []*big.Int, []float64, error) {}
func (e *EthAPIBackend) BlobBaseFee(ctx context.Context) *big.Int {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) ExtRPCEnabled() bool {
	return true
}

func (e *EthAPIBackend) RPCGasCap() uint64 {
	return 50000000
}

func (e *EthAPIBackend) RPCEVMTimeout() time.Duration {
	return 5 * time.Second
}

func (e *EthAPIBackend) RPCTxFeeCap() float64 {
	return 1
}

func (e *EthAPIBackend) UnprotectedAllowed() bool {
	return e.allowUnprotectedTxs
}

func (e *EthAPIBackend) SetHead(number uint64) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, *yutypes.Header, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("headerByNumber").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("headerByNumber").Inc()
	var (
		yuBlock *yutypes.CompactBlock
		err     error
	)
	switch number {
	case rpc.PendingBlockNumber:
		// FIXME
		yuBlock, err = e.chain.Chain.GetEndCompactBlock()
	case rpc.LatestBlockNumber, rpc.FinalizedBlockNumber, rpc.SafeBlockNumber:
		yuBlock, err = e.chain.Chain.LastFinalizedCompact()
	default:
		yuBlock, err = e.chain.Chain.GetCompactBlockByHeight(yucommon.BlockNum(number))
	}

	if yuBlock == nil {
		return nil, nil, err
	}
	return yuHeader2EthHeader(yuBlock.Header), yuBlock.Header, err
}

func (e *EthAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, *yutypes.Header, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("headerByHash").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("headerByHash").Inc()
	yuBlock, err := e.chain.Chain.GetCompactBlock(yucommon.Hash(hash))
	if err != nil {
		logrus.Error("ethrpc.api_backend.HeaderByHash() failed: ", err)
		return nil, nil, err
	}

	return yuHeader2EthHeader(yuBlock.Header), yuBlock.Header, err
}

func (e *EthAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, *yutypes.Header, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("headerByNumberOrHash").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("headerByNumberOrHash").Inc()
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return e.HeaderByNumber(ctx, blockNr)
	}

	if blockHash, ok := blockNrOrHash.Hash(); ok {
		return e.HeaderByHash(ctx, blockHash)
	}

	return nil, nil, errors.New("invalid arguments; neither block number nor hash specified")
}

func (e *EthAPIBackend) CurrentHeader() *types.Header {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("currentHeader").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("currentHeader").Inc()
	yuBlock, err := e.chain.Chain.GetEndCompactBlock()
	if err != nil {
		logrus.Error("EthAPIBackend.CurrentBlock() failed: ", err)
		return nil
	}

	return yuHeader2EthHeader(yuBlock.Header)
}

func (e *EthAPIBackend) CurrentBlock() *types.Header {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("currentBlock").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("currentBlock").Inc()
	yuBlock, err := e.chain.Chain.GetEndCompactBlock()
	if err != nil {
		logrus.Error("EthAPIBackend.CurrentBlock() failed: ", err)
		return nil
	}

	return yuHeader2EthHeader(yuBlock.Header)
}

func (e *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, *yutypes.Block, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("blockByNumber").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("blockByNumber").Inc()
	var (
		yuBlock *yutypes.Block
		err     error
	)
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		switch number {
		case rpc.PendingBlockNumber:
			yuBlock, err = e.chain.Chain.GetEndBlock()
		case rpc.LatestBlockNumber:
			yuBlock, err = e.chain.Chain.GetEndBlock()
		case rpc.FinalizedBlockNumber, rpc.SafeBlockNumber:
			yuBlock, err = e.chain.Chain.LastFinalized()
		default:
			yuBlock, err = e.chain.Chain.GetBlockByHeight(yucommon.BlockNum(number))
		}

		if err == nil {
			break
		}

		logrus.Warnf("BlockByNumber attempt %d failed: blockNumber=%v, error=%v", attempt, number, err)
		if attempt < MaxRetries {
			time.Sleep(RetryIntervalMs)
		}
	}
	if err != nil {
		logrus.Errorf("rpc BlockByNumber failed: blockNumber=%v, error=%v", number, err)
		return nil, nil, err
	}
	block, err := e.compactBlock2EthBlock(yuBlock)
	if err != nil {
		logrus.Errorf("BlockByNumber failed to convert compact block: blockNumber=%v, error=%v", number, err)
		return nil, nil, err
	}
	return block, yuBlock, err
}

func (e *EthAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, *yutypes.Block, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("blockByHash").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("blockByHash").Inc()
	yuBlock, err := e.chain.Chain.GetBlock(yucommon.Hash(hash))
	if err != nil {
		return nil, nil, err
	}
	block, err := e.compactBlock2EthBlock(yuBlock)
	if err != nil {
		return nil, nil, err
	}
	return block, yuBlock, err
}

func (e *EthAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, *yutypes.Block, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("blockByNumberOrHash").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("blockByNumberOrHash").Inc()
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return e.BlockByNumber(ctx, blockNr)
	}

	if blockHash, ok := blockNrOrHash.Hash(); ok {
		return e.BlockByHash(ctx, blockHash)
	}

	return nil, nil, errors.New("invalid arguments; neither block number nor hash specified")
}

func (e *EthAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("stateAndHeaderByNumber").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("stateAndHeaderByNumber").Inc()
	header, _, err := e.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	tri := e.chain.GetTripodInstance(SolidityTripod)
	solidityTri := tri.(*evm.Solidity)
	stateDB, err := solidityTri.StateAt(header.Root)
	if err != nil {
		return nil, nil, err
	}
	return stateDB, header, nil
}

func (e *EthAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("stateAndHeaderByNumberOrHash").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("stateAndHeaderByNumberOrHash").Inc()
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return e.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		yuBlock, err := e.chain.Chain.GetBlock(yucommon.Hash(hash))
		if err != nil {
			return nil, nil, err
		}
		tri := e.chain.GetTripodInstance(SolidityTripod)
		solidityTri := tri.(*evm.Solidity)
		stateDB, err := solidityTri.StateAt(common.Hash(yuBlock.StateRoot))
		if err != nil {
			return nil, nil, err
		}
		return stateDB, yuHeader2EthHeader(yuBlock.Header), nil
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (e *EthAPIBackend) ChainDb() ethdb.Database {
	EthApiBackendCounter.WithLabelValues("chainDb").Inc()
	tri := e.chain.GetTripodInstance(SolidityTripod)
	solidityTri := tri.(*evm.Solidity)
	ethDB := solidityTri.GetEthDB()
	return ethDB
}

func (e *EthAPIBackend) AccountManager() *accounts.Manager {
	// TODO implement me
	return nil
}

func (e *EthAPIBackend) Pending() (*types.Block, types.Receipts, *state.StateDB) {
	// TODO implement me
	panic("implement me")
}

// Eth has changed to POS, Td(total difficulty) is for POW
func (e *EthAPIBackend) GetTd(ctx context.Context, hash common.Hash) *big.Int {
	return nil
}

func (e *EthAPIBackend) GetEVM(ctx context.Context, msg *core.Message, state *state.StateDB, header *types.Header, vmConfig *vm.Config, blockCtx *vm.BlockContext) *vm.EVM {
	EthApiBackendCounter.WithLabelValues("getEVM").Inc()
	if vmConfig == nil {
		// vmConfig = e.chain.Chain.GetVMConfig()
		vmConfig = &vm.Config{
			EnablePreimageRecording: false, // TODO: replace with ctx.Bool()
		}
	}
	txContext := core.NewEVMTxContext(msg)
	var context vm.BlockContext
	if blockCtx != nil {
		context = *blockCtx
	} else {
		var b Backend
		context = core.NewEVMBlockContext(header, NewChainContext(ctx, b), nil)
	}
	return vm.NewEVM(context, txContext, state, e.ChainConfig(), *vmConfig)
}

func (e *EthAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) Call(ctx context.Context, args TransactionArgs, blockNrOrHash *rpc.BlockNumberOrHash, overrides *StateOverride, blockOverrides *BlockOverrides) (hexutil.Bytes, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("call").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("call").Inc()
	globalGasCap := e.RPCGasCap()
	if blockNrOrHash == nil {
		latest := rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)
		blockNrOrHash = &latest
	}
	stateDb, header, err := e.StateAndHeaderByNumberOrHash(ctx, *blockNrOrHash)
	if stateDb == nil || err != nil {
		return nil, err
	}

	if err := args.CallDefaults(globalGasCap, header.BaseFee, e.ChainConfig().ChainID); err != nil {
		return nil, err
	}

	if args.To == nil {
		return nil, errors.New("missing 'to' in params")
	}

	callRequest := evm.CallRequest{
		Address:  *args.To,
		Input:    args.data(),
		Value:    args.Value.ToInt(),
		GasLimit: uint64(*args.Gas),
		GasPrice: args.GasPrice.ToInt(),
	}
	callRequest.Origin = args.from()

	requestByt, _ := json.Marshal(callRequest)
	rdCall := new(yucommon.RdCall)
	rdCall.TripodName = SolidityTripod
	rdCall.FuncName = "Call"
	rdCall.Params = string(requestByt)

	response, err := e.chain.HandleRead(rdCall)
	if err != nil {
		return nil, err
	}

	resp := response.DataInterface.(*evm.CallResponse)
	return resp.Ret, nil
}

func (e *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("sendTx").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("sendTx").Inc()
	// Check if this tx has been created
	signedTxHash := signedTx.Hash()
	exist, _, _, _, _, err := e.GetTransaction(ctx, signedTxHash)
	if err != nil {
		logrus.Errorf("[SendTx] Failed to get transaction, txHash(%s), yuHash(%s), error: %v", signedTxHash.Hex(), yucommon.Hash(signedTxHash).Hex(), err)
		return err
	}
	if exist {
		return errors.Errorf("tx(%s) already known onchain", signedTxHash.String())
	}
	existedTx, err := e.GetPoolTransaction(signedTxHash)
	if err != nil {
		logrus.Errorf("[SendTx] Failed to get transaction from txpool, txHash(%s), yuHash(%s), error: %v", signedTxHash.Hex(), yucommon.Hash(signedTxHash).Hex(), err)
		return err
	}
	if existedTx != nil {
		return errors.Errorf("tx(%s) already known in txpool", signedTxHash.String())
	}

	// Create Tx
	head := e.CurrentBlock()
	signer := types.MakeSigner(e.ChainConfig(), head.Number, head.Time)
	sender, err := types.Sender(signer, signedTx)
	if err != nil {
		logrus.Errorf("[SendTx] Failed to get sender, txHash(%s), yuHash(%s), error: %v", signedTxHash.Hex(), yucommon.Hash(signedTxHash).Hex(), err)
		return err
	}
	v, r, s := signedTx.RawSignatureValues()
	txArg := NewTxArgsFromTx(signedTx)
	txArgByte, _ := json.Marshal(txArg)
	txReq := &evm.TxRequest{
		Input:    signedTx.Data(),
		Origin:   sender,
		Address:  signedTx.To(),
		GasLimit: signedTx.Gas(),
		GasPrice: signedTx.GasPrice(),
		Value:    signedTx.Value(),
		Hash:     signedTx.Hash(),
		Nonce:    signedTx.Nonce(),
		V:        v,
		R:        r,
		S:        s,

		OriginArgs: txArgByte,
	}
	byt, err := json.Marshal(txReq)
	if err != nil {
		logrus.Errorf("[SendTx] Failed to marshal txReq, txHash(%s), yuHash(%s), error: %v", signedTxHash.Hex(), yucommon.Hash(signedTxHash).Hex(), err)
		return err
	}
	signedWrCall := &protocol.SignedWrCall{
		Call: &yucommon.WrCall{
			TripodName: SolidityTripod,
			FuncName:   "ExecuteTxn",
			Params:     string(byt),
		},
	}
	return e.chain.HandleTxn(signedWrCall)
}

func YuTxn2EthTxn(yuSignedTxn *yutypes.SignedTxn) (*types.Transaction, error) {
	// Un-serialize wrCall.params to retrieve data:
	wrCallParams := yuSignedTxn.Raw.WrCall.Params
	txReq := &evm.TxRequest{}
	err := json.Unmarshal([]byte(wrCallParams), txReq)
	if err != nil {
		return nil, err
	}

	// if nonce is assigned to signedTx.Raw.Nonce, then this is ok; otherwise it's nil:
	txArgs := &TransactionArgs{}
	err = json.Unmarshal(txReq.OriginArgs, txArgs)
	if err != nil {
		return nil, err
	}
	tx := txArgs.ToTransaction(txReq.V, txReq.R, txReq.S)
	return tx, nil
}

func (e *EthAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (bool, *types.Transaction, common.Hash, uint64, uint64, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getTransaction").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getTransaction").Inc()
	// Used to get txn from either txdb & txpool:
	stxn, err := e.chain.GetTxn(yucommon.Hash(txHash))
	if err != nil {
		logrus.Errorf("[GetTransaction] Failed to get transaction from txdb or txpool, txHash(%s), yuHash(%s), error: %v", txHash.Hex(), yucommon.Hash(txHash).Hex(), err)
		return false, nil, common.Hash{}, 0, 0, err
	}
	if stxn == nil {
		logrus.Debugf("[GetTransaction] Transaction not found, txHash(%s), yuHash(%s)", txHash.Hex(), yucommon.Hash(txHash).Hex())
		return false, nil, common.Hash{}, 0, 0, nil
	}
	ethTxn, err := YuTxn2EthTxn(stxn)
	if err != nil {
		logrus.Errorf("[GetTransaction] Failed to convert transaction, txHash(%s), yuHash(%s), error: %v", txHash.Hex(), yucommon.Hash(txHash).Hex(), err)
		return false, nil, common.Hash{}, 0, 0, err
	}

	// rcptReq := &evm.ReceiptRequest{Hash: txHash}
	receipt, err := e.chain.TxDB.GetReceipt(yucommon.Hash(txHash))
	if err != nil {
		logrus.Errorf("[GetTransaction] Failed to get receipt from txdb, txHash(%s), yuHash(%s), error: %v", txHash.Hex(), yucommon.Hash(txHash).Hex(), err)
		return false, nil, common.Hash{}, 0, 0, err
	}
	if receipt == nil {
		logrus.Debugf("[GetTransaction] Receipt not found, txHash(%s), yuHash(%s)", txHash.Hex(), yucommon.Hash(txHash).Hex())
		return false, nil, common.Hash{}, 0, 0, nil
	}

	//resp, err := e.adaptChainRead(rcptReq, "GetReceipt")
	//if err != nil {
	//	return false, nil, common.Hash{}, 0, 0, err
	//}
	//receiptResponse := resp.DataInterface.(*evm.ReceiptResponse)
	//if receiptResponse.Err != nil {
	//	return false, nil, common.Hash{}, 0, 0, errors.Errorf("StatusCode: %d, Error: %v", resp.StatusCode, receiptResponse.Err)
	//}
	//receipt := receiptResponse.Receipt

	blockHash := receipt.BlockHash
	blockNumber := receipt.Height
	var index uint64
	if receipt.Extra != nil {
		ethRcpt := new(types.Receipt)
		err = json.Unmarshal(receipt.Extra, ethRcpt)
		if err != nil {
			logrus.Error("GetTransaction() json.Unmarshal eth receipt failed: ", err)
			return true, ethTxn, common.Hash(blockHash), uint64(blockNumber), 0, err
		}
		index = uint64(ethRcpt.TransactionIndex)
	}

	return true, ethTxn, common.Hash(blockHash), uint64(blockNumber), index, nil
}

func (e *EthAPIBackend) GetReceiptsForLog(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getReceiptsForLog").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getReceiptsForLog").Inc()
	compactBlock, err := e.chain.Chain.GetCompactBlock(yucommon.Hash(blockHash))
	if err != nil {
		return nil, err
	}
	var receipts []*types.Receipt
	for _, txHash := range compactBlock.TxnsHashes {
		rcptReq := &evm.ReceiptRequest{Hash: common.Hash(txHash)}
		resp, err := e.adaptChainRead(rcptReq, "GetReceipt")
		if err != nil {
			continue
		}
		receiptResponse := resp.DataInterface.(*evm.ReceiptResponse)
		if receiptResponse.Err != nil {
			continue
		}
		receipts = append(receipts, receiptResponse.Receipt)
		// if compactBlock.Height == 182 {
		// 	fmt.Println("receiptResponse.Receipt TxHash: ", receiptResponse.Receipt.TxHash)
		// }
	}
	return receipts, nil
}

func (e *EthAPIBackend) GetReceipt(ctx context.Context, txnHash common.Hash) (*types.Receipt, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("GetReceipt").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("GetReceipt").Inc()
	rcptReq := &evm.ReceiptRequest{Hash: txnHash}
	resp, err := e.adaptChainRead(rcptReq, "GetReceipt")
	if err != nil {
		return nil, err
	}
	receiptsResponse := resp.DataInterface.(*evm.ReceiptResponse)
	if receiptsResponse.Err != nil {
		return nil, errors.Errorf("StatusCode: %d, Error: %v", resp.StatusCode, receiptsResponse.Err)
	}
	return receiptsResponse.Receipt, nil
}

func (e *EthAPIBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getReceipts").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getReceipts").Inc()
	compactBlock, err := e.chain.Chain.GetCompactBlock(yucommon.Hash(blockHash))
	if err != nil {
		return nil, err
	}

	yuTxHashes := compactBlock.TxnsHashes
	//trans yuTxHashes to common.Hash
	txHashes := make([]common.Hash, len(yuTxHashes))
	for i, yuTxHash := range yuTxHashes {
		txHashes[i] = common.Hash(yuTxHash)
	}
	if len(txHashes) == 0 {
		return nil, nil
	}

	rcptReq := &evm.ReceiptsRequest{Hashes: txHashes}
	resp, err := e.adaptChainRead(rcptReq, "GetReceipts")
	if err != nil {
		return nil, err
	}
	receiptsResponse := resp.DataInterface.(*evm.ReceiptsResponse)
	if receiptsResponse.Err != nil {
		return nil, errors.Errorf("StatusCode: %d, Error: %v", resp.StatusCode, receiptsResponse.Err)
	}

	return receiptsResponse.Receipts, nil
}

func (e *EthAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getPoolTransactions").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getPoolTransactions").Inc()
	// Similar to: e.chain.ChainEnv.Pool.GetTxn - ChainEnv can be ignored b/c txpool has index based on hxHash, therefore it's unique
	stxn, _ := e.chain.Pool.GetAllTxns() // will not return error here

	var ethTxns []*types.Transaction

	for _, yuSignedTxn := range stxn {
		ethTxn, err := YuTxn2EthTxn(yuSignedTxn)
		if err != nil {
			return nil, err
		}
		ethTxns = append(ethTxns, ethTxn)
	}

	return ethTxns, nil
}

// Similar to GetTransaction():
func (e *EthAPIBackend) GetPoolTransaction(txHash common.Hash) (*types.Transaction, error) {
	stxn, err := e.chain.Pool.GetTxn(yucommon.Hash(txHash)) // will not return error here
	if err != nil || stxn == nil {
		return nil, err
	}

	return YuTxn2EthTxn(stxn)
}

func (e *EthAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getPoolNonce").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getPoolNonce").Inc()
	// Loop through all transactions to find matching Account Address, and return it's nonce (if have)
	allEthTxns, _ := e.GetPoolTransactions()

	head := e.CurrentBlock()
	signer := types.MakeSigner(e.ChainConfig(), head.Number, head.Time)

	nonce := uint64(0)
	for _, ethTxn := range allEthTxns {
		sender, _ := types.Sender(signer, ethTxn)
		if sender == addr {
			nonce++
			// return ethTxn.Nonce(), nil
		}
	}

	return nonce, nil
}

func (e *EthAPIBackend) Stats() (pending int, queued int) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) TxPoolContent() (map[common.Address][]*types.Transaction, map[common.Address][]*types.Transaction) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) TxPoolContentFrom(addr common.Address) ([]*types.Transaction, []*types.Transaction) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) SubscribeNewTxsEvent(events chan<- core.NewTxsEvent) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) ChainConfig() *params.ChainConfig {
	return e.ethChainCfg
}

func (e *EthAPIBackend) Engine() consensus.Engine {
	return FakeEngine{}
}

func (e *EthAPIBackend) GetBody(ctx context.Context, hash common.Hash, number rpc.BlockNumber) (*types.Body, error) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) GetLogs(ctx context.Context, blockHash common.Hash, number uint64) ([][]*types.Log, error) {
	start := time.Now()
	defer func() {
		EthApiBackendDuration.WithLabelValues("getLogs").Observe(float64(time.Since(start).Microseconds()))
	}()
	EthApiBackendCounter.WithLabelValues("getLogs").Inc()
	if blockHash == (common.Hash{}) {
		_, yuHeader, _ := e.HeaderByNumber(ctx, rpc.BlockNumber(number))
		blockHash = common.Hash(yuHeader.Hash)
	}
	receipts, err := e.GetReceiptsForLog(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	result := [][]*types.Log{}
	for _, receipt := range receipts {
		logs := []*types.Log{}
		for _, vLog := range receipt.Logs {
			logs = append(logs, vLog)
		}
		result = append(result, logs)
	}

	return result, nil
}

func (e *EthAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) BloomStatus() (uint64, uint64) {
	// TODO implement me
	panic("implement me")
}

func (e *EthAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	// TODO implement me
	panic("implement me")
}

func yuHeader2EthHeader(yuHeader *yutypes.Header) *types.Header {
	return &types.Header{
		ParentHash:  common.Hash(yuHeader.PrevHash),
		Coinbase:    common.Address{}, // FIXME
		Root:        common.Hash(yuHeader.StateRoot),
		TxHash:      common.Hash(yuHeader.TxnRoot),
		ReceiptHash: common.Hash(yuHeader.ReceiptRoot),
		Difficulty:  new(big.Int).SetUint64(yuHeader.Difficulty),
		Number:      new(big.Int).SetUint64(uint64(yuHeader.Height)),
		GasLimit:    yuHeader.LeiLimit,
		GasUsed:     yuHeader.LeiUsed,
		Time:        yuHeader.Timestamp,
		Extra:       yuHeader.Extra,
		Nonce:       types.BlockNonce{},
		BaseFee:     big.NewInt(params.InitialBaseFee),
	}
}

func (e *EthAPIBackend) compactBlock2EthBlock(yuBlock *yutypes.Block) (*types.Block, error) {
	header := yuHeader2EthHeader(yuBlock.Header)

	// Generate transactions and receipts
	var ethTxs []*types.Transaction
	var txHashes []common.Hash
	for _, yuSignedTxn := range yuBlock.Txns {
		tx, err := YuTxn2EthTxn(yuSignedTxn)
		if err != nil {
			return nil, err
		}
		ethTxs = append(ethTxs, tx)
		txHashes = append(txHashes, tx.Hash())
	}

	var receipts []*types.Receipt
	rcptReq := &evm.ReceiptsRequest{Hashes: txHashes}
	resp, err := e.adaptChainRead(rcptReq, "GetReceipts")
	if err != nil {
		logrus.Errorf("Failed to get receipts when adaptChainRead: %v", err)
	} else {
		receiptResponse := resp.DataInterface.(*evm.ReceiptsResponse)
		if receiptResponse.Err != nil {
			logrus.Errorf("Failed to get receipts when compact block: error-code: %d, error: %v", resp.StatusCode, receiptResponse.Err)
		} else {
			receipts = receiptResponse.Receipts
		}
	}

	return types.NewBlock(header, ethTxs, nil, receipts, trie.NewStackTrie(nil)), nil
}

func (e *EthAPIBackend) adaptChainRead(req any, funcName string) (*yucontext.ResponseData, error) {
	byt, err := json.Marshal(req)
	if err != nil {
		logrus.Error(fmt.Errorf("EthAPIBackend %v meet err: %v", funcName, err))
		return nil, err
	}
	params := string(byt)

	rdCall := &yucommon.RdCall{
		TripodName: SolidityTripod,
		FuncName:   funcName,
		Params:     params,
	}

	resp, err := e.chain.HandleRead(rdCall)
	if err != nil {
		logrus.Error(fmt.Errorf("EthAPIBackend %v meet err: %v, param:%v", funcName, err, params))
		return nil, fmt.Errorf("EthAPIBackend %v meet err: %v", funcName, err)
	}
	return resp, nil
}

// region ---- Fake Consensus Engine ----

type FakeEngine struct{}

// Author retrieves the Ethereum address of the account that minted the given block.
func (f FakeEngine) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (f FakeEngine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	panic("Unimplemented fake engine method VerifyHeader")
}

// VerifyHeaders checks whether a batch of headers conforms to the consensus rules.
func (f FakeEngine) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	panic("Unimplemented fake engine method VerifyHeaders")
}

// VerifyUncles verifies that the given block's uncles conform to the consensus rules.
func (f FakeEngine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	panic("Unimplemented fake engine method VerifyUncles")
}

// Prepare initializes the consensus fields of a block header.
func (f FakeEngine) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	panic("Unimplemented fake engine method Prepare")
}

// Finalize runs any post-transaction state modifications.
func (f FakeEngine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, body *types.Body) {
	panic("Unimplemented fake engine method Finalize")
}

// FinalizeAndAssemble runs any post-transaction state modifications and assembles the final block.
func (f FakeEngine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, body *types.Body, receipts []*types.Receipt) (*types.Block, error) {
	panic("Unimplemented fake engine method FinalizeAndAssemble")
}

// Seal generates a new sealing request for the given input block.
func (f FakeEngine) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	panic("Unimplemented fake engine method Seal")
}

// SealHash returns the hash of a block prior to it being sealed.
func (f FakeEngine) SealHash(header *types.Header) common.Hash {
	panic("Unimplemented fake engine method SealHash")
}

// CalcDifficulty is the difficulty adjustment algorithm.
func (f FakeEngine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	panic("Unimplemented fake engine method CalcDifficulty")
}

// APIs returns the RPC APIs this consensus engine provides.
func (f FakeEngine) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	panic("Unimplemented fake engine method APIs")
}

// Close terminates any background threads maintained by the consensus engine.
func (f FakeEngine) Close() error {
	panic("Unimplemented fake engine method Close")
}

// endregion  ---- Fake Consensus Engine ----
