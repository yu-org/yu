package sql

import (
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage"
	"gorm.io/gorm"
)

type SqlDB interface {
	storage.StorageType
	Db() *gorm.DB
	CreateIfNotExist(table interface{}) error
	AutoMigrate(table any) error
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
