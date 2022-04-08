package config

type StateConf struct {
	KV MptKvConf `toml:"kv"`
}

type MptKvConf struct {
	IndexDB  KVconf `toml:"index_db"`
	NodeBase KVconf `toml:"node_base"`
}

type EvmKvConf struct {
	IndexDB  KVconf `toml:"index_db"`
	NodeBase KVconf `toml:"node_base"`

	// evm raw leveldb
	Fpath     string ` toml:"fpath"`
	Cache     int    ` toml:"cache"`
	Handles   int    ` toml:"handles"`
	Namespace string ` toml:"namespace"`
	ReadOnly  bool   ` toml:"read_only"`
}
