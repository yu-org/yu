package config

type KVconf struct {
	// "bolt"
	KvType string `toml:"kv_type"`
	// dbpath, such as boltdb, pebble
	Path string `toml:"path"`
	// distributed kvdb
	Hosts []string `toml:"hosts"`

	UseSQlDbConf bool      `toml:"use_sql_db"`
	SQLDbConf    SqlDbConf `toml:"sql_db"`
}

type SqlDbConf struct {
	SqlDbType          string `toml:"sql_db_type"`
	Dsn                string `toml:"dsn"`
	MaxOpenConnections int    `toml:"max_open_connections"`
	MaxIdleConnections int    `toml:"max_idle_connections"`
}

type SqliteDBConf struct {
	Path string `toml:"path"`
}

type TxnConf struct {
	UseTxnConf          bool `toml:"use_txn_conf"`
	EnableSqliteStorage bool `toml:"enable_sqlite_storage"`
}
