package sql

import (
	"github.com/yu-altar/yu/config"
	"github.com/yu-altar/yu/storage"
	"github.com/yu-altar/yu/yerror"
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
