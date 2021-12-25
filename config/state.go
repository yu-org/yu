package config

type StateConf struct {
	KV StateKvConf `toml:"kv"`
}

type StateKvConf struct {
	IndexDB  KVconf `toml:"index_db"`
	NodeBase KVconf `toml:"node_base"`
}

type StateEvmConf struct {
	Fpath     string ` toml:"fpath"`
	Cache     int    ` toml:"cache"`
	Handles   int    ` toml:"handles"`
	Namespace string ` toml:"namespace"`
	ReadOnly  bool   ` toml:"read_only"`
}
