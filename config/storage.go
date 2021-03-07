package config

type KVconf struct {
	// "bolt" "badger" "tikv"
	KvType string `toml:"kv_type"`
	// dbpath, such as boltdb, pebble
	Path string `toml:"path"`
	// distributed kvdb
	Hosts []string `toml:"hosts"`
}

type QueueConf struct {
	QueueType string `toml:"queue_type"`
	Url       string `toml:"url"`
	// json, gob, default
	Encoder string `toml:"encoder"`
}
