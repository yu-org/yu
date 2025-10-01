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
	IsReddioMainnet bool `toml:"is_reddio_mainnet"`

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

	// EventsWatcher configs
	EnableBridge               bool     `toml:"enable_bridge"`
	L1ClientAddress            string   `toml:"l1_client_address"`
	L2ClientAddress            string   `toml:"l2_client_address"`
	ParentLayerContractAddress string   `toml:"parentlayer_contract_address"`
	ChildLayerContractAddress  string   `toml:"childlayer_contract_address"`
	L2BlockCollectionDepth     *big.Int `toml:"l2_block_collection_depth"`
	BridgeHost                 string   `toml:"bridge_host"`
	BridgePort                 string   `toml:"bridge_port"`
	// BridgeDBConfig             *database.Config `toml:"bridge_db_config"`
	// watcher config
	L1WatcherConfig BridgeWatcherConfig `toml:"l1_watcher_config"`
	L2WatcherConfig BridgeWatcherConfig `toml:"l2_watcher_config"`
	// relayer config
	RelayerBatchSize            int    `toml:"relayer_batch_size"`
	MultisigEnvFile             string `toml:"multisig_env_file"`
	MultisigEnvVar              string `toml:"multisig_env_var"`
	RelayerEnvFile              string `toml:"relayer_env_file"`
	RelayerEnvVar               string `toml:"relayer_env_var"`
	L1_RawBridgeEventsTableName string `toml:"l1_raw_bridge_events_table_name"`
	L2_RawBridgeEventsTableName string `toml:"l2_raw_bridge_events_table_name"`

	// checker config
	EnableBridgeChecker bool                `toml:"enable_bridge_checker"`
	BridgeCheckerConfig BridgeCheckerConfig `toml:"bridge_checker_config"`
}
type BridgeWatcherConfig struct {
	Confirmation uint64 `toml:"confirmation"`
	FetchLimit   uint64 `toml:"fetch_limit"`
	StartHeight  uint64 `toml:"start_height"`
	BlockTime    uint64 `toml:"block_time"`
	ChainID      int64  `toml:"chain_id"`
}
type BridgeCheckerConfig struct {
	CheckerBatchSize       int    `toml:"checker_batch_size"`
	SepoliaTickerInterval  int    `toml:"sepolia_ticker_interval"`
	ReddioTickerInterval   int    `toml:"reddio_ticker_interval"`
	EnableL1CheckStep1     bool   `toml:"enable_l1_check_step1"`
	EnableL1CheckStep2     bool   `toml:"enable_l1_check_step2"`
	EnableL2CheckStep1     bool   `toml:"enable_l2_check_step1"`
	EnableL2CheckStep2     bool   `toml:"enable_l2_check_step2"`
	CheckL1ContractAddress string `toml:"check_l1_contract_address"`
	CheckL2ContractAddress string `toml:"check_l2_contract_address"`
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
