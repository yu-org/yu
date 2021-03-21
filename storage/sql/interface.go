package sql

import (
	"yu/config"
	"yu/storage"
	"yu/yerror"
)

type SqlDB interface {
	storage.StorageType
}

func NewSqlDB(cfg *config.SqlDbConf) (SqlDB, error) {
	switch cfg.SqlDbType {
	case "sqlite":
		return NewSqlite(cfg.Dsn)
	default:
		return nil, yerror.NoSqlDbType
	}
}
