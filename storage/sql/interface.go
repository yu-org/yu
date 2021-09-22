package sql

import (
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/storage"
	"github.com/yu-org/yu/yerror"
	"gorm.io/gorm"
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
