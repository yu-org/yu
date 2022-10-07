package config

type KVconf struct {
	// "bolt"
	KvType string `toml:"kv_type"`
	// dbpath, such as boltdb, pebble
	Path string `toml:"path"`
	// distributed kvdb
	Hosts []string `toml:"hosts"`
}

type QueueConf struct {
	QueueType string `toml:"queue_type"`
	Url       string `toml:"url"`
	// toml, gob, default
	Encoder string `toml:"encoder"`
}

type SqlDbConf struct {
	SqlDbType string `toml:"sql_db_type"`
	Dsn       string `toml:"dsn"`
}
