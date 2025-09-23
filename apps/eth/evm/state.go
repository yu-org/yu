package evm

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/ethereum/go-ethereum/triedb/pathdb"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"

	"github.com/yu-org/yu/apps/eth/config"
)

type EthState struct {
	cfg        *config.Config
	stateDB    *state.StateDB
	stateCache state.Database
	trieDB     *triedb.Database
	ethDB      ethdb.Database
	snaps      *snapshot.Tree
	logger     *tracing.Hooks
}

func NewEthState(cfg *config.Config, currentStateRoot common.Hash) (*EthState, error) {
	vmConfig := vm.Config{
		EnablePreimageRecording: cfg.EnablePreimageRecording,
	}
	if cfg.VMTrace != "" {
		var traceConfig json.RawMessage
		if cfg.VMTraceConfig != "" {
			traceConfig = json.RawMessage(cfg.VMTraceConfig)
		}
		t, err := tracers.LiveDirectory.New(cfg.VMTrace, traceConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracer %s: %v", cfg.VMTrace, err)
		}
		vmConfig.Tracer = t
	}

	db, err := rawdb.Open(memorydb.New(), rawdb.OpenOptions{
		Ancient:          filepath.Join(cfg.DbPath, cfg.Ancient),
		MetricsNamespace: cfg.NameSpace,
		ReadOnly:         false,
	})
	if err != nil {
		return nil, err
	}

	//cacheCfg, err := cacheConfig(cfg, db)
	//if err != nil {
	//	return nil, err
	//}
	snapCfg := snapsConfig(cfg)

	trieDB := triedb.NewDatabase(db, trieConfig(false))

	snaps, err := snapshot.New(snapCfg, db, trieDB, currentStateRoot)
	if err != nil {
		return nil, err
	}
	stateCache := state.NewDatabase(trieDB, snaps)
	stateDB, _ := state.New(types.EmptyRootHash, stateCache)

	ethState := &EthState{
		cfg:        cfg,
		stateDB:    stateDB,
		stateCache: stateCache,
		trieDB:     trieDB,
		ethDB:      db,
		snaps:      snaps,
		logger:     vmConfig.Tracer,
	}
	err = ethState.newStateForNextBlock(currentStateRoot)
	return ethState, err
}

func (s *EthState) StateDB() *state.StateDB {
	return s.stateDB
}

func (s *EthState) SetStateDB(d *state.StateDB) {
	s.stateDB = d
}

func (s *EthState) setTxContext(txHash common.Hash, txIdx int) {
	s.stateDB.SetTxContext(txHash, txIdx)
}

func (s *EthState) StateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, s.stateCache)
}

func (s *EthState) GenesisCommit() (common.Hash, error) {
	return s.Commit(0)
}

//func (s *EthState) NewStateDB(parentStateRoot common.Hash) error {
//	statedb, err := state.New(parentStateRoot, s.stateCache, s.snaps)
//	if err != nil {
//		return err
//	}
//	statedb.SetLogger(s.logger)
//	// Enable prefetching to pull in trie node paths while processing transactions
//	statedb.StartPrefetcher("chain")
//	s.StateDB = statedb
//	return err
//}

func (s *EthState) Commit(blockNum uint64) (common.Hash, error) {
	s.stateDB.StopPrefetcher()
	stateRoot, err := s.stateDB.Commit(blockNum, true, true)
	if err != nil {
		return common.Hash{}, err
	}
	err = s.trieDB.Commit(stateRoot, true)
	if err != nil {
		return common.Hash{}, err
	}

	// new stateDB for the next block
	err = s.newStateForNextBlock(stateRoot)
	logrus.Debug("EthState Commit Successful")
	return stateRoot, err
}

func (s *EthState) newStateForNextBlock(currentStateRoot common.Hash) error {
	// new stateDB for the next block
	newsStateDB, err := state.New(currentStateRoot, s.stateCache)
	if err != nil {
		return err
	}
	// newsStateDB.SetLogger(s.logger)
	// Enable prefetching to pull in trie node paths while processing transactions
	newsStateDB.StartPrefetcher("chain", nil, nil)
	s.stateDB = newsStateDB
	return nil
}

func trieConfig(isVerkle bool) *triedb.Config {
	cfg := core.DefaultConfig()
	config := &triedb.Config{
		Preimages: cfg.Preimages,
		IsVerkle:  isVerkle,
	}
	if cfg.StateScheme == rawdb.HashScheme {
		config.HashDB = &hashdb.Config{
			CleanCacheSize: cfg.TrieCleanLimit * 1024 * 1024,
		}
	}
	if cfg.StateScheme == rawdb.PathScheme {
		config.PathDB = &pathdb.Config{
			StateHistory:        cfg.StateHistory,
			EnableStateIndexing: cfg.ArchiveMode,
			TrieCleanSize:       cfg.TrieCleanLimit * 1024 * 1024,
			StateCleanSize:      cfg.SnapshotLimit * 1024 * 1024,
			JournalDirectory:    cfg.TrieJournalDirectory,

			// TODO(rjl493456442): The write buffer represents the memory limit used
			// for flushing both trie data and state data to disk. The config name
			// should be updated to eliminate the confusion.
			WriteBufferSize: cfg.TrieDirtyLimit * 1024 * 1024,
			NoAsyncFlush:    cfg.TrieNoAsyncFlush,
		}
	}
	return config
}

func snapsConfig(cfg *config.Config) snapshot.Config {
	return snapshot.Config{
		CacheSize:  cfg.SnapshotCache,
		Recovery:   cfg.Recovery,
		NoBuild:    cfg.NoBuild,
		AsyncBuild: !cfg.SnapshotWait,
	}
}

func (s *EthState) Prepare(rules params.Rules, sender, coinbase common.Address, dst *common.Address, precompiles []common.Address, list types.AccessList) {
	s.stateDB.StopPrefetcher()
	s.stateDB.Prepare(rules, sender, coinbase, dst, precompiles, list)
}

//func (s *EthState) SetNonce(addr common.Address, nonce uint64) {
//	s.stateDB.StopPrefetcher()
//	s.stateDB.SetNonce(addr, nonce)
//}

func (s *EthState) GetNonce(addr common.Address) uint64 {
	s.stateDB.StopPrefetcher()
	uint64 := s.stateDB.GetNonce(addr)
	return uint64
}

func (s *EthState) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	s.stateDB.AddBalance(addr, amount, reason)
}

func (s *EthState) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	s.stateDB.SubBalance(addr, amount, reason)
}
