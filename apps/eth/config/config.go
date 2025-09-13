package config

import "time"

type Config struct {
	// Enables VM tracing
	VMTrace       string `toml:"vm_trace"`
	VMTraceConfig string `toml:"vm_trace_config"`
	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool `toml:"enable_preimage_recording"`

	// snapshots config
	Recovery      bool `toml:"recovery"`
	NoBuild       bool `toml:"no_build"`
	SnapshotWait  bool `toml:"snapshot_wait"`
	SnapshotCache int  `toml:"snapshot_cache"`

	// cache config
	TrieCleanCache int           `toml:"trie_clean_cache"`
	TrieDirtyCache int           `toml:"trie_dirty_cache"`
	TrieTimeout    time.Duration `toml:"trie_timeout"`
	Preimages      bool          `toml:"preimages"`
	NoPruning      bool          `toml:"no_pruning"`
	NoPrefetch     bool          `toml:"no_prefetch"`
	StateHistory   uint64        `toml:"state_history"`

	StateScheme string `toml:"state_scheme"`

	// database
	DbPath    string `toml:"db_path"`
	DbType    string `toml:"db_type"`    // pebble
	NameSpace string `toml:"name_space"` // eth/db/chaindata/
	Ancient   string `toml:"ancient"`
	Cache     int    `toml:"cache"`
	Handles   int    `toml:"handles"`
}
