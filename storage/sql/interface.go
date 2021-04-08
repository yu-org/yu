package sql

import (
	"gorm.io/gorm"
	"yu/config"
	"yu/storage"
	"yu/yerror"
)

type SqlDB interface {
	storage.StorageType
	Db() *gorm.DB
	CreateIfNotExist(table interface{}) error
}

func NewSqlDB(cfg *config.SqlDbConf) (SqlDB, error) {
	switch cfg.SqlDbType {
	case "sqlite":
		return NewSqlite(cfg.Dsn)
	case "mysql":
		return NewMysql(cfg.Dsn)
	case "postgre":
		return NewPostgreSql(cfg.Dsn)
	default:
		return nil, yerror.NoSqlDbType
	}
}
