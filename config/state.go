package config

type StateConf struct {
	KV StateKvConf `toml:"kv"`
}

type StateKvConf struct {
	IndexDB  KVconf `toml:"index_db"`
	NodeBase KVconf `toml:"node_base"`
}
