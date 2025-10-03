package config

import (
	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	"github.com/sirupsen/logrus"
	"math/big"
	"time"
)

type GethConfig struct {
	IsMainnet bool `toml:"is_mainnet"`

	ChainConfig *params.ChainConfig

	// BlockContext provides the EVM with auxiliary information. Once provided
	// it shouldn't be modified.
	GetHashFn   func(n uint64) common.Hash
	Coinbase    common.Address
	GasLimit    uint64
	BlockNumber *big.Int
	Time        uint64
	Difficulty  *big.Int
	BaseFee     *big.Int
	BlobBaseFee *big.Int
	Random      *common.Hash

	// TxContext provides the EVM with information about a transaction.
	// All fields can change between transactions.
	Origin     common.Address
	GasPrice   *big.Int
	BlobHashes []common.Hash
	BlobFeeCap *big.Int

	// Unknown
	Value     *big.Int
	Debug     bool
	EVMConfig vm.Config

	// Global config
	EnableEthRPC bool   `toml:"enable_eth_rpc"`
	EthHost      string `toml:"eth_host"`
	EthPort      string `toml:"eth_port"`

	// chainID
	ChainID int64 `toml:"chain_id"`
}

func (gc *GethConfig) Copy() *GethConfig {
	return &GethConfig{
		ChainConfig:  gc.ChainConfig,
		GetHashFn:    gc.GetHashFn,
		Coinbase:     gc.Coinbase,
		GasLimit:     gc.GasLimit,
		BlockNumber:  gc.BlockNumber,
		Time:         gc.Time,
		Difficulty:   gc.Difficulty,
		BaseFee:      gc.BaseFee,
		BlobBaseFee:  gc.BlobBaseFee,
		Random:       gc.Random,
		Origin:       gc.Origin,
		GasPrice:     gc.GasPrice,
		BlobHashes:   gc.BlobHashes,
		BlobFeeCap:   gc.BlobFeeCap,
		Value:        gc.Value,
		Debug:        gc.Debug,
		EVMConfig:    gc.EVMConfig,
		EnableEthRPC: gc.EnableEthRPC,
		EthHost:      gc.EthHost,
		EthPort:      gc.EthPort,
	}
}

// sets defaults on the config
func SetDefaultGethConfig(fpath string) *GethConfig {
	cfg := &GethConfig{
		ChainConfig: params.AllEthashProtocolChanges,
		Difficulty:  big.NewInt(1),
		Origin:      common.HexToAddress("0x0"),
		Coinbase:    common.HexToAddress("0x3E2D75F83e775761890d9ab9389eCF6C9D6017eB"),
		BlockNumber: big.NewInt(0),
		Time:        0,
		GasLimit:    8000000,
		GasPrice:    big.NewInt(1),
		Value:       big.NewInt(0),
		Debug:       false,
		EVMConfig:   vm.Config{},
		BaseFee:     big.NewInt(params.InitialBaseFee), // 1 gwei
		BlobBaseFee: big.NewInt(params.BlobTxMinBlobGasprice),
		BlobHashes:  []common.Hash{},
		BlobFeeCap:  big.NewInt(0),
		Random:      &common.Hash{},
		GetHashFn: func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		},
		ChainID: 50341,
	}
	_, err := toml.DecodeFile(fpath, cfg)
	if err != nil {
		logrus.Fatalf("load config file failed: %v", err)
	}
	cfg.ChainConfig.ChainID = big.NewInt(cfg.ChainID)

	return cfg
}

func LoadGethConfig(fpath string) *GethConfig {
	cfg := SetDefaultGethConfig(fpath)
	return cfg
}

func SetDefaultEthStateConfig() *Config {
	return &Config{
		VMTrace:                 "",
		VMTraceConfig:           "",
		EnablePreimageRecording: false,
		Recovery:                false,
		NoBuild:                 false,
		SnapshotWait:            false,
		SnapshotCache:           128,              // Default cache size
		TrieCleanCache:          256,              // Default Trie cleanup cache size
		TrieDirtyCache:          256,              // Default Trie dirty cache size
		TrieTimeout:             60 * time.Second, // Default Trie timeout
		Preimages:               false,
		NoPruning:               false,
		NoPrefetch:              false,
		StateHistory:            0,                   // By default, there is no state history
		StateScheme:             "hash",              // Default state scheme
		DbPath:                  "yu_eth_db",         // Default database path
		DbType:                  "pebble",            // Default database type
		NameSpace:               "eth/db/chaindata/", // Default namespace
		Ancient:                 "ancient",           // Default ancient data path
		Cache:                   512,                 // Default cache size
		Handles:                 64,                  // Default number of handles
	}
}
