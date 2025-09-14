package ethrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	yutypes "github.com/yu-org/yu/core/types"
)

var (
	errInvalidTopic           = errors.New("invalid topic(s)")
	errFilterNotFound         = errors.New("filter not found")
	errInvalidBlockRange      = errors.New("invalid block range params")
	errPendingLogsUnsupported = errors.New("pending logs are not supported")
	errExceedMaxTopics        = errors.New("exceed max topics")
)

const (
	maxTopics          = 100
	maxSubTopics       = 1000
	getLogsLimitBlocks = 500
)

// FilterCriteria represents a request to create a new filter.
// Same as ethereum.FilterQuery but with UnmarshalJSON() method.
type FilterCriteria ethereum.FilterQuery

// UnmarshalJSON sets *args fields with given data.
func (args *FilterCriteria) UnmarshalJSON(data []byte) error {
	type input struct {
		BlockHash *common.Hash     `json:"blockHash"`
		FromBlock *rpc.BlockNumber `json:"fromBlock"`
		ToBlock   *rpc.BlockNumber `json:"toBlock"`
		Addresses interface{}      `json:"address"`
		Topics    []interface{}    `json:"topics"`
	}

	var raw input
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.BlockHash != nil {
		if raw.FromBlock != nil || raw.ToBlock != nil {
			// BlockHash is mutually exclusive with FromBlock/ToBlock criteria
			return errors.New("cannot specify both BlockHash and FromBlock/ToBlock, choose one or the other")
		}
		args.BlockHash = raw.BlockHash
	} else {
		if raw.FromBlock != nil {
			args.FromBlock = big.NewInt(raw.FromBlock.Int64())
		}

		if raw.ToBlock != nil {
			args.ToBlock = big.NewInt(raw.ToBlock.Int64())
		}
	}

	args.Addresses = []common.Address{}

	if raw.Addresses != nil {
		// raw.Address can contain a single address or an array of addresses
		switch rawAddr := raw.Addresses.(type) {
		case []interface{}:
			for i, addr := range rawAddr {
				if strAddr, ok := addr.(string); ok {
					addr, err := decodeAddress(strAddr)
					if err != nil {
						return fmt.Errorf("invalid address at index %d: %v", i, err)
					}
					args.Addresses = append(args.Addresses, addr)
				} else {
					return fmt.Errorf("non-string address at index %d", i)
				}
			}
		case string:
			addr, err := decodeAddress(rawAddr)
			if err != nil {
				return fmt.Errorf("invalid address: %v", err)
			}
			args.Addresses = []common.Address{addr}
		default:
			return errors.New("invalid addresses in query")
		}
	}
	//if len(raw.Topics) > maxTopics {
	//	return errExceedMaxTopics
	//}

	// topics is an array consisting of strings and/or arrays of strings.
	// JSON null values are converted to common.Hash{} and ignored by the filter manager.
	if len(raw.Topics) > 0 {
		args.Topics = make([][]common.Hash, len(raw.Topics))
		for i, t := range raw.Topics {
			switch topic := t.(type) {
			case nil:
				// ignore topic when matching logs

			case string:
				// match specific topic
				top, err := decodeTopic(topic)
				if err != nil {
					return err
				}
				args.Topics[i] = []common.Hash{top}

			case []interface{}:
				// or case e.g. [null, "topic0", "topic1"]
				//if len(topic) > maxSubTopics {
				//	return errExceedMaxTopics
				//}
				for _, rawTopic := range topic {
					if rawTopic == nil {
						// null component, match all
						args.Topics[i] = nil
						break
					}
					if topic, ok := rawTopic.(string); ok {
						parsed, err := decodeTopic(topic)
						if err != nil {
							return err
						}
						args.Topics[i] = append(args.Topics[i], parsed)
					} else {
						return errInvalidTopic
					}
				}
			default:
				return errInvalidTopic
			}
		}
	}

	return nil
}

func decodeAddress(s string) (common.Address, error) {
	b, err := hexutil.Decode(s)
	if err == nil && len(b) != common.AddressLength {
		err = fmt.Errorf("hex has invalid length %d after decoding; expected %d for address", len(b), common.AddressLength)
	}
	return common.BytesToAddress(b), err
}

func decodeTopic(s string) (common.Hash, error) {
	b, err := hexutil.Decode(s)
	if err == nil && len(b) != common.HashLength {
		err = fmt.Errorf("hex has invalid length %d after decoding; expected %d for topic", len(b), common.HashLength)
	}
	return common.BytesToHash(b), err
}

type LogFilter struct {
	b Backend

	addresses []common.Address
	topics    [][]common.Hash

	block      *common.Hash // Block hash if filtering a single block
	begin, end int64        // Range interval if filtering multiple blocks
}

func newLogFilter(ctx context.Context, b Backend, crit FilterCriteria) (*LogFilter, error) {
	var filter *LogFilter
	if crit.BlockHash != nil {
		filter = &LogFilter{
			b:         b,
			block:     crit.BlockHash,
			addresses: crit.Addresses,
			topics:    crit.Topics,
		}
	} else {
		begin := rpc.LatestBlockNumber.Int64()
		if crit.FromBlock != nil {
			begin = crit.FromBlock.Int64()
		}

		end := rpc.LatestBlockNumber.Int64()
		if crit.ToBlock != nil {
			end = crit.ToBlock.Int64()
		}

		if begin == rpc.PendingBlockNumber.Int64() || end == rpc.PendingBlockNumber.Int64() {
			return nil, errPendingLogsUnsupported
		}

		_, hdr, _ := b.HeaderByNumber(ctx, rpc.LatestBlockNumber)
		if begin == rpc.LatestBlockNumber.Int64() {
			begin = int64(hdr.Height)
		}
		if end == rpc.LatestBlockNumber.Int64() {
			end = int64(hdr.Height)
		}

		if begin > 0 && end > 0 && begin > end {
			return nil, errInvalidBlockRange
		}
		if end-begin > getLogsLimitBlocks {
			return nil, errors.New("block range is too wide")
		}

		filter = &LogFilter{
			b:         b,
			begin:     begin,
			end:       end,
			addresses: crit.Addresses,
			topics:    crit.Topics,
		}
	}

	return filter, nil
}

func (f *LogFilter) Logs(ctx context.Context) ([]*types.Log, error) {
	if f.block != nil {
		_, yuHeader, err := f.b.HeaderByHash(ctx, *f.block)
		if err != nil {
			return nil, err
		}
		if yuHeader == nil {
			return nil, errors.New("unknown block")
		}
		return f.FilterLogs(ctx, yuHeader)
	} else {
		var result []*types.Log
		for ; f.begin <= f.end; f.begin++ {
			_, yuHeader, err := f.b.HeaderByNumber(ctx, rpc.BlockNumber(f.begin))
			if err != nil {
				logrus.Errorf("[GetLog] Failed to getHeaderByNumber %v, error: %s", f.begin, err)
				return nil, err
			}
			logs, err := f.FilterLogs(ctx, yuHeader)
			if err != nil {
				return nil, err
			}
			result = append(result, logs...)
		}
		return result, nil
	}
}

func (f *LogFilter) FilterLogs(ctx context.Context, yuHeader *yutypes.Header) ([]*types.Log, error) {
	logs, err := f.b.GetLogs(ctx, common.Hash(yuHeader.Hash), uint64(yuHeader.Height))
	if err != nil {
		return nil, err
	}
	result := make([]*types.Log, 0)
	var logIdx uint
	for i, txLogs := range logs {
		for _, vLog := range txLogs {
			vLog.BlockHash = common.Hash(yuHeader.Hash)
			vLog.BlockNumber = uint64(yuHeader.Height)
			vLog.TxIndex = uint(i)
			vLog.Index = logIdx
			logIdx++

			if f.checkMatches(ctx, vLog) {
				result = append(result, vLog)
			}
		}
	}
	return result, nil
}

func (f *LogFilter) checkMatches(ctx context.Context, vLog *types.Log) bool {
	if len(f.addresses) > 0 {
		if !slices.Contains(f.addresses, vLog.Address) {
			return false
		}
	}

	// TODO: The logic for topic filtering is a bit complex; it will not be implemented for now.
	if len(f.topics) > len(vLog.Topics) {
		return false
	}

	for i, sub := range f.topics {
		if len(sub) == 0 {
			continue // empty rule set == wildcard
		}

		if !slices.Contains(sub, vLog.Topics[i]) {

			return false
		}
	}

	return true
}

// rangeLogsAsync retrieves block-range logs that match the filter criteria asynchronously,
// it creates and returns two channels: one for delivering log data, and one for reporting errors.
func (f *LogFilter) rangeLogsAsync(ctx context.Context) (chan *types.Log, chan error) {
	var (
		logChan = make(chan *types.Log)
		errChan = make(chan error)
	)

	go func() {
		defer func() {
			close(errChan)
			close(logChan)
		}()

		// Gather all indexed logs, and finish with non indexed ones
		var (
			end = uint64(f.end)
			// size, sections = f.sys.backend.BloomStatus()
			// err            error
		)
		// if indexed := sections * size; indexed > uint64(f.begin) {
		// 	if indexed > end {
		// 		indexed = end + 1
		// 	}
		// 	if err = f.indexedLogs(ctx, indexed-1, logChan); err != nil {
		// 		errChan <- err
		// 		return
		// 	}
		// }

		if err := f.unindexedLogs(ctx, end, logChan); err != nil {
			errChan <- err
			return
		}

		errChan <- nil
	}()

	return logChan, errChan
}

// unindexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *LogFilter) unindexedLogs(ctx context.Context, end uint64, logChan chan *types.Log) error {
	for ; f.begin <= int64(end); f.begin++ {
		header, _, err := f.b.HeaderByNumber(ctx, rpc.BlockNumber(f.begin))
		if header == nil || err != nil {
			return err
		}
		found, err := f.blockLogs(ctx, header)
		if err != nil {
			return err
		}
		for _, log := range found {
			select {
			case logChan <- log:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}
func (f *LogFilter) blockLogs(ctx context.Context, header *types.Header) ([]*types.Log, error) {
	//if bloomFilter(header.Bloom, f.addresses, f.topics) {
	return f.checkMatchesFromHeader(ctx, header)
	//}
	//return nil, nil
}

// checkMatches checks if the receipts belonging to the given header contain any log events that
// match the filter criteria. This function is called when the bloom filter signals a potential match.
// skipFilter signals all logs of the given block are requested.
func (f *LogFilter) checkMatchesFromHeader(ctx context.Context, header *types.Header) ([]*types.Log, error) {
	// hash := header.Hash()
	// // Logs in cache are partially filled with context data
	// // such as tx index, block hash, etc.
	// // Notably tx hash is NOT filled in because it needs
	// // access to block body data.
	// cached, err := f.sys.cachedLogElem(ctx, hash, header.Number.Uint64())
	// if err != nil {
	// 	return nil, err
	// }
	// logs := filterLogs(cached.logs, nil, nil, f.addresses, f.topics)
	// if len(logs) == 0 {
	// 	return nil, nil
	// }
	// // Most backends will deliver un-derived logs, but check nevertheless.
	// if len(logs) > 0 && logs[0].TxHash != (common.Hash{}) {
	// 	return logs, nil
	// }

	// body, err := f.sys.cachedGetBody(ctx, cached, hash, header.Number.Uint64())
	// if err != nil {
	// 	return nil, err
	// }
	// for i, log := range logs {
	// 	// Copy log not to modify cache elements
	// 	logcopy := *log
	// 	logcopy.TxHash = body.Transactions[logcopy.TxIndex].Hash()
	// 	logs[i] = &logcopy
	// }
	// return logs, nil
	return nil, nil
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*types.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*types.Log {
	var check = func(log *types.Log) bool {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			return false
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			return false
		}
		if len(addresses) > 0 && !slices.Contains(addresses, log.Address) {
			return false
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			return false
		}
		for i, sub := range topics {
			if len(sub) == 0 {
				continue // empty rule set == wildcard
			}
			if !slices.Contains(sub, log.Topics[i]) {
				return false
			}
		}
		return true
	}
	var ret []*types.Log
	for _, log := range logs {
		if check(log) {
			ret = append(ret, log)
		}
	}
	return ret
}
