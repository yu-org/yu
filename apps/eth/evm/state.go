package evm

import (
	"math"

	"math/big"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"

	"github.com/yu-org/yu/apps/eth/config"
	yu_types "github.com/yu-org/yu/core/types"
)

type EthState struct {
	db        ethdb.Database   // Pebble-backed ethdb
	trieDB    *triedb.Database // triedb database
	stateDB   *state.StateDB   // state DB used by EVM
	cachingDB *state.CachingDB

	chainCfg *params.ChainConfig // chain config for EVM
	gethCfg  *config.GethConfig  // geth config for EVM

	root common.Hash // current state root
}

func NewEthState(root common.Hash, gethCfg *config.GethConfig) (*EthState, error) {
	dbPath := filepath.Join("ethstate", "chaindata")
	pdb, err := pebble.New(dbPath, 128, 16, "", false)
	if err != nil {
		return nil, err
	}

	db, err := rawdb.Open(pdb, rawdb.OpenOptions{Ancient: filepath.Join(dbPath, "ancient")})
	if err != nil {
		return nil, err
	}

	trieDB := triedb.NewDatabase(db, &triedb.Config{Preimages: true})

	snapCfg := snapshot.Config{CacheSize: 512}
	snapObj, err := snapshot.New(snapCfg, db, trieDB, root)
	if err != nil {
		db.Close()
		return nil, err
	}

	cachingDB := state.NewDatabase(trieDB, snapObj)
	sdb, err := state.New(root, cachingDB)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &EthState{
		db:        db,
		trieDB:    trieDB,
		stateDB:   sdb,
		cachingDB: cachingDB,
		chainCfg:  gethCfg.ChainConfig,
		gethCfg:   gethCfg,
		root:      root,
	}, nil
}

func (s *EthState) StateAt(root common.Hash) (*state.StateDB, error) {
	return state.New(root, s.cachingDB)
}

func (s *EthState) GenesisCommit(genesis *Genesis) (common.Hash, error) {
	// return flushAlloc(&genesis.Alloc, s.trieDB)

	root, err := flushAlloc(&genesis.Alloc, s.trieDB)
	if err != nil {
		return common.Hash{}, err
	}
	s.root = root
	return root, nil
}

func (s *EthState) StartState() error {
	statedb, err := state.New(s.root, s.cachingDB)
	if err != nil {
		return err
	}
	s.stateDB = statedb

	return nil
}

func (s *EthState) ApplyTx(block *yu_types.Block, tx *ethtypes.Transaction, txIdx int, gp *core.GasPool, usedGas *uint64) (*ethtypes.Receipt, error) {
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     s.gethCfg.GetHashFn,
		Coinbase:    s.gethCfg.Coinbase,
		BlockNumber: s.gethCfg.BlockNumber,
		Time:        s.gethCfg.Time,
		Difficulty:  s.gethCfg.Difficulty,
		GasLimit:    tx.Gas(),
		BaseFee:     s.gethCfg.BaseFee,
		BlobBaseFee: s.gethCfg.BlobBaseFee,
		Random:      s.gethCfg.Random,
	}

	evm := vm.NewEVM(blockContext, s.stateDB, s.gethCfg.ChainConfig, s.gethCfg.EVMConfig)

	msg, err := core.TransactionToMessage(tx, ethtypes.MakeSigner(s.chainCfg, block.Height.ToBigInt(), block.Timestamp), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	s.stateDB.SetTxContext(common.Hash(block.Hash), txIdx)
	rcpt, err := core.ApplyTransactionWithEVM(msg, gp, s.stateDB, block.Height.ToBigInt(), common.Hash(block.Hash), block.Timestamp, tx, usedGas, evm)
	if err != nil {
		return nil, err
	}
	if rcpt.Logs == nil {
		rcpt.Logs = []*ethtypes.Log{}
	}
	return rcpt, nil
}

func (s *EthState) ApplyTxForReader(msg *core.Message) (*core.ExecutionResult, error) {
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     s.gethCfg.GetHashFn,
		Coinbase:    s.gethCfg.Coinbase,
		BlockNumber: s.gethCfg.BlockNumber,
		Time:        s.gethCfg.Time,
	}
	evm := vm.NewEVM(blockContext, s.stateDB, s.gethCfg.ChainConfig, s.gethCfg.EVMConfig)

	gp := new(core.GasPool).AddGas(math.MaxUint64)
	return core.ApplyMessage(evm, msg, gp)
}

func (s *EthState) Commit(blockNum uint64) (common.Hash, error) {
	root, err := s.stateDB.Commit(blockNum, true, false)
	if err != nil {
		return common.Hash{}, err
	}
	if err := s.trieDB.Commit(root, true); err != nil {
		return common.Hash{}, err
	}
	s.root = root
	// s.blockNum = blockNum
	return root, nil
}
