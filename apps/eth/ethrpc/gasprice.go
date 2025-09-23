package ethrpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"

	//"github.com/yu-org/yu/common"
	"github.com/ethereum/go-ethereum/common"
)

const sampleNumber = 3 // Number of transactions sampled in a block

// Oracle recommends gas prices based on the content of recent
// blocks. Suitable for both light and full clients.
type EthGasPrice struct {
	backend     Backend
	lastHead    common.Hash
	lastPrice   *big.Int
	maxPrice    *big.Int
	ignorePrice *big.Int
	// cacheLock   sync.RWMutex
	// fetchLock   sync.Mutex

	checkBlocks, percentile int
	// maxHeaderHistory, maxBlockHistory uint64

	// historyCache *lru.Cache[cacheKey, processedFees]
}

func NewEthGasPrice(backend Backend) *EthGasPrice {
	// default value from geth->config.go->FullNodeGPO
	return &EthGasPrice{
		backend:     backend,
		checkBlocks: 20,
		percentile:  60,
		maxPrice:    big.NewInt(500 * params.GWei),
		ignorePrice: big.NewInt(2 * params.Wei),
		lastPrice:   big.NewInt(params.GWei), // lastPrice default value from geth->miner.go->DefaultConfig
	}
}

// SuggestTipCap returns a tip cap so that newly created transaction can have a
// very high chance to be included in the following blocks.
//
// Note, for legacy transactions and the legacy eth_gasPrice RPC call, it will be
// necessary to add the basefee to the returned number to fall back to the legacy
// behavior.
func (e *EthAPIBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	//head, _, _ := e.HeaderByNumber(ctx, rpc.LatestBlockNumber)
	//headHash := head.Hash()
	//
	//// TODO: need add cache lock
	//// If the latest gasprice is still available, return it.
	//// e.cacheLock.RLock()
	//lastHead, lastPrice := e.gasPriceCache.lastHead, e.gasPriceCache.lastPrice
	//// oracle.cacheLock.RUnlock()
	//if headHash == lastHead {
	//	return new(big.Int).Set(lastPrice), nil
	//}
	//// oracle.fetchLock.Lock()
	//// defer oracle.fetchLock.Unlock()
	//
	//// Try checking the cache again, maybe the last fetch fetched what we need
	////oracle.cacheLock.RLock()
	////lastHead, lastPrice = ethGasPrice.lastHead, ethGasPrice.lastPrice
	////oracle.cacheLock.RUnlock()
	////if headHash == lastHead {
	////	return new(big.Int).Set(lastPrice), nil
	////}
	//var (
	//	sent, exp int
	//	number    = head.Number.Uint64()
	//	result    = make(chan results, e.gasPriceCache.checkBlocks)
	//	quit      = make(chan struct{})
	//	results   []*big.Int
	//)
	//for sent < e.gasPriceCache.checkBlocks && number > 0 {
	//	go e.getBlockValues(ctx, number, sampleNumber, e.gasPriceCache.ignorePrice, result, quit)
	//	sent++
	//	exp++
	//	number--
	//}
	//for exp > 0 {
	//	res := <-result
	//	if res.err != nil {
	//		close(quit)
	//		return new(big.Int).Set(lastPrice), res.err
	//	}
	//	exp--
	//	// Nothing returned. There are two special cases here:
	//	// - The block is empty
	//	// - All the transactions included are sent by the miner itself.
	//	// In these cases, use the latest calculated price for sampling.
	//	if len(res.values) == 0 {
	//		res.values = []*big.Int{lastPrice}
	//	}
	//	// Besides, in order to collect enough data for sampling, if nothing
	//	// meaningful returned, try to query more blocks. But the maximum
	//	// is 2*checkBlocks.
	//	if len(res.values) == 1 && len(results)+1+exp < e.gasPriceCache.checkBlocks*2 && number > 0 {
	//		go e.getBlockValues(ctx, number, sampleNumber, e.gasPriceCache.ignorePrice, result, quit)
	//		sent++
	//		exp++
	//		number--
	//	}
	//	results = append(results, res.values...)
	//}
	//price := lastPrice
	//if len(results) > 0 {
	//	slices.SortFunc(results, func(a, b *big.Int) int { return a.Cmp(b) })
	//	price = results[(len(results)-1)*e.gasPriceCache.percentile/100]
	//}
	//if price.Cmp(e.gasPriceCache.maxPrice) > 0 {
	//	price = new(big.Int).Set(e.gasPriceCache.maxPrice)
	//}
	//// oracle.cacheLock.Lock()
	//e.gasPriceCache.lastHead = headHash
	//e.gasPriceCache.lastPrice = price
	//// oracle.cacheLock.Unlock()
	//
	//return new(big.Int).Set(price), nil
	return new(big.Int).SetUint64(1), nil
}

type results struct {
	values []*big.Int
	err    error
}

// getBlockValues calculates the lowest transaction gas price in a given block
// and sends it to the result channel. If the block is empty or all transactions
// are sent by the miner itself(it doesn't make any sense to include this kind of
// transaction prices for sampling), nil gasprice is returned.
func (e *EthAPIBackend) getBlockValues(ctx context.Context, blockNum uint64, limit int, ignoreUnder *big.Int, result chan results, quit chan struct{}) {
	block, _, err := e.BlockByNumber(ctx, rpc.BlockNumber(blockNum))
	if block == nil {
		select {
		case result <- results{nil, err}:
		case <-quit:
		}
		return
	}
	signer := types.MakeSigner(e.ChainConfig(), block.Number(), block.Time())

	// Sort the transaction by effective tip in ascending sort.
	txs := block.Transactions()
	sortedTxs := make([]*types.Transaction, len(txs))
	copy(sortedTxs, txs)
	baseFee := block.BaseFee()
	slices.SortFunc(sortedTxs, func(a, b *types.Transaction) int {
		// It's okay to discard the error because a tx would never be
		// accepted into a block with an invalid effective tip.
		tip1, _ := a.EffectiveGasTip(baseFee)
		tip2, _ := b.EffectiveGasTip(baseFee)
		return tip1.Cmp(tip2)
	})

	var prices []*big.Int
	for _, tx := range sortedTxs {
		tip, _ := tx.EffectiveGasTip(baseFee)
		if ignoreUnder != nil && tip.Cmp(ignoreUnder) == -1 {
			continue
		}
		sender, err := types.Sender(signer, tx)
		if err == nil && sender != block.Coinbase() {
			prices = append(prices, tip)
			if len(prices) >= limit {
				break
			}
		}
	}
	select {
	case result <- results{prices, nil}:
	case <-quit:
	}
}

func (e *EthAPIBackend) FeeHistory(ctx context.Context, blockCount uint64, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (*big.Int, [][]*big.Int, []*big.Int, []float64, []*big.Int, []float64, error) {
	if blockCount < 1 {
		return common.Big0, nil, nil, nil, nil, nil, nil
	}

	oldestBlock := common.Big0
	currentHeader := e.CurrentHeader().Number.Uint64()
	resolvedLastBlock := uint64(0)
	if lastBlock < 0 {
		switch lastBlock {
		case rpc.PendingBlockNumber:
			// TODO: Don't know how to implement it
		case rpc.LatestBlockNumber:
			// Retrieved above.
			resolvedLastBlock = currentHeader
		case rpc.SafeBlockNumber:
			header, _, _ := e.HeaderByNumber(ctx, rpc.SafeBlockNumber)
			resolvedLastBlock = header.Number.Uint64()
		case rpc.FinalizedBlockNumber:
			header, _, _ := e.HeaderByNumber(ctx, rpc.FinalizedBlockNumber)
			resolvedLastBlock = header.Number.Uint64()
		case rpc.EarliestBlockNumber:
			header, _, _ := e.HeaderByNumber(ctx, rpc.EarliestBlockNumber)
			resolvedLastBlock = header.Number.Uint64()
		}
	} else {
		resolvedLastBlock = uint64(lastBlock)
	}

	if blockCount <= resolvedLastBlock {
		oldestBlock = big.NewInt(int64(resolvedLastBlock + 1 - blockCount))
	}

	results := make([]*blockFees, 0)

	for i := 0; i < int(blockCount); i++ {
		blockNumber := oldestBlock
		if blockNumber.Uint64() > resolvedLastBlock {
			break
		}
		fees := &blockFees{blockNumber: blockNumber.Uint64()}

		if len(rewardPercentiles) > 0 {
			fees.block, _, fees.err = e.BlockByNumber(ctx, rpc.BlockNumber(blockNumber.Int64()))
			if fees.block != nil && fees.err != nil {
				fees.receipts, fees.err = e.GetReceipts(ctx, fees.block.Hash())
				fees.header = fees.block.Header()
			}
		} else {
			fees.header, _, fees.err = e.HeaderByNumber(ctx, rpc.BlockNumber(blockNumber.Int64()))
		}

		if fees.header != nil && fees.err == nil {
			// process block
			e.processBlock(fees, rewardPercentiles)
		}

		results = append(results, fees)
	}

	var (
		reward           = make([][]*big.Int, blockCount)
		baseFee          = make([]*big.Int, blockCount+1)
		gasUsedRatio     = make([]float64, blockCount)
		blobGasUsedRatio = make([]float64, blockCount)
		blobBaseFee      = make([]*big.Int, blockCount+1)
		firstMissing     = blockCount
	)

	for i, fees := range results {
		if fees.err != nil {
			return common.Big0, nil, nil, nil, nil, nil, fees.err
		}
		if fees.results.baseFee != nil {
			rewardIndex := fees.blockNumber - oldestBlock.Uint64()
			reward[rewardIndex], baseFee[rewardIndex], baseFee[rewardIndex+1], gasUsedRatio[rewardIndex] = fees.results.reward, fees.results.baseFee, fees.results.nextBaseFee, fees.results.gasUsedRatio
			blobGasUsedRatio[rewardIndex], blobBaseFee[rewardIndex], blobBaseFee[rewardIndex+1] = fees.results.blobGasUsedRatio, fees.results.blobBaseFee, fees.results.nextBlobBaseFee
		} else {
			// 如果没有block和error，意味着我们请求到了未来的区块（可能因为重组）
			if uint64(i) < firstMissing {
				firstMissing = uint64(i)
			}
		}
	}

	if firstMissing == 0 {
		return common.Big0, nil, nil, nil, nil, nil, nil
	}
	if len(rewardPercentiles) != 0 {
		reward = reward[:firstMissing]
	} else {
		reward = nil
	}
	baseFee, gasUsedRatio = baseFee[:firstMissing+1], gasUsedRatio[:firstMissing]
	blobBaseFee, blobGasUsedRatio = blobBaseFee[:firstMissing+1], blobGasUsedRatio[:firstMissing]
	return oldestBlock, reward, baseFee, gasUsedRatio, blobBaseFee, blobGasUsedRatio, nil
}

// blockFees represents a single block for processing
type blockFees struct {
	// set by the caller
	blockNumber uint64
	header      *types.Header
	block       *types.Block
	receipts    types.Receipts
	results     processedFees

	err error
}

type processedFees struct {
	reward                       []*big.Int
	baseFee, nextBaseFee         *big.Int
	gasUsedRatio                 float64
	blobGasUsedRatio             float64
	blobBaseFee, nextBlobBaseFee *big.Int
}

// txGasAndReward is sorted in ascending order based on reward
type txGasAndReward struct {
	gasUsed uint64
	reward  *big.Int
}

// resolveBlockRange resolves the specified block range to absolute block numbers while also
// enforcing backend specific limitations. The pending block and corresponding receipts are
// also returned if requested and available.
// Note: an error is only returned if retrieving the head header has failed. If there are no
// retrievable blocks in the specified range then zero block count is returned with no error.
func (e *EthAPIBackend) resolveBlockRange(ctx context.Context, reqEnd rpc.BlockNumber, blocks uint64) (*types.Block, []*types.Receipt, uint64, uint64, error) {
	var (
		headBlock       *types.Header
		pendingBlock    *types.Block
		pendingReceipts types.Receipts
		err             error
	)

	// Get the chain's current head.
	if headBlock, _, err = e.HeaderByNumber(ctx, rpc.LatestBlockNumber); err != nil {
		return nil, nil, 0, 0, err
	}
	head := rpc.BlockNumber(headBlock.Number.Uint64())
	// Fail if request block is beyond the chain's current head.
	if head < reqEnd {
		return nil, nil, 0, 0, fmt.Errorf("%w: requested %d, head %d", errors.New("request beyond head block"), reqEnd, head)
	}

	if reqEnd < 0 {
		var (
			resolved *types.Header
			err      error
		)
		switch reqEnd {
		case rpc.PendingBlockNumber:
			if pendingBlock, pendingReceipts, _ = e.Pending(); pendingBlock != nil {
				resolved = pendingBlock.Header()
			} else {
				// Pending block not supported by backend, process only until latest block.
				resolved = headBlock

				// Update total blocks to return to account for this.
				blocks--
			}
		case rpc.LatestBlockNumber:
			// Retrieved above.
			resolved = headBlock
		case rpc.SafeBlockNumber:
			resolved, _, err = e.HeaderByNumber(ctx, rpc.SafeBlockNumber)
		case rpc.FinalizedBlockNumber:
			resolved, _, err = e.HeaderByNumber(ctx, rpc.FinalizedBlockNumber)
		case rpc.EarliestBlockNumber:
			resolved, _, err = e.HeaderByNumber(ctx, rpc.EarliestBlockNumber)
		}
		if resolved == nil || err != nil {
			return nil, nil, 0, 0, err
		}
		// Absolute number resolved.
		reqEnd = rpc.BlockNumber(resolved.Number.Uint64())
	}

	// If there are no blocks to return, short circuit.
	if blocks == 0 {
		return nil, nil, 0, 0, nil
	}
	// Ensure not trying to retrieve before genesis.
	if uint64(reqEnd+1) < blocks {
		blocks = uint64(reqEnd + 1)
	}
	return pendingBlock, pendingReceipts, uint64(reqEnd), blocks, nil
}

func (e *EthAPIBackend) processBlock(bf *blockFees, percentiles []float64) {
	config := e.ChainConfig()

	if bf.results.baseFee = bf.header.BaseFee; bf.results.baseFee == nil {
		bf.results.baseFee = new(big.Int)
	}

	bf.results.nextBaseFee = eip1559.CalcBaseFee(config, bf.header)

	if excessBlobGas := bf.header.ExcessBlobGas; excessBlobGas != nil {
		bf.results.blobBaseFee = eip4844.CalcBlobFee(*excessBlobGas)
		bf.results.nextBlobBaseFee = eip4844.CalcBlobFee(eip4844.CalcExcessBlobGas(*excessBlobGas, *bf.header.BlobGasUsed))
	} else {
		bf.results.blobBaseFee = new(big.Int)
		bf.results.nextBlobBaseFee = new(big.Int)
	}

	bf.results.gasUsedRatio = float64(bf.header.GasUsed) / float64(bf.header.GasLimit)
	if blobGasUsed := bf.header.BlobGasUsed; blobGasUsed != nil {
		bf.results.blobGasUsedRatio = float64(*blobGasUsed) / params.MaxBlobGasPerBlock
	}

	if len(percentiles) == 0 {
		return
	}

	if bf.block == nil || (bf.receipts == nil && len(bf.block.Transactions()) != 0) {
		logrus.Error("Block or receipts are missing while reward percentiles are requested")
		return
	}

	bf.results.reward = make([]*big.Int, len(percentiles))
	if len(bf.block.Transactions()) == 0 {
		// return an all zero row if there are no transactions to gather data from
		for i := range bf.results.reward {
			bf.results.reward[i] = new(big.Int)
		}
		return
	}

	sorter := make([]txGasAndReward, len(bf.block.Transactions()))
	for i, tx := range bf.block.Transactions() {
		reward, _ := tx.EffectiveGasTip(bf.block.BaseFee())
		sorter[i] = txGasAndReward{gasUsed: bf.receipts[i].GasUsed, reward: reward}
	}
	slices.SortStableFunc(sorter, func(a, b txGasAndReward) int {
		return a.reward.Cmp(b.reward)
	})

	var txIndex int
	sumGasUsed := sorter[0].gasUsed

	for i, p := range percentiles {
		thresholdGasUsed := uint64(float64(bf.block.GasUsed()) * p / 100)
		for sumGasUsed < thresholdGasUsed && txIndex < len(bf.block.Transactions())-1 {
			txIndex++
			sumGasUsed += sorter[txIndex].gasUsed
		}
		bf.results.reward[i] = sorter[txIndex].reward
	}
}
