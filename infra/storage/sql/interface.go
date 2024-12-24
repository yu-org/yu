package sql

import (
	"gorm.io/gorm"

	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage"
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
		return NewSqlite(cfg)
	case "mysql":
		return NewMysql(cfg)
	case "postgre":
		return NewPostgreSql(cfg)
	default:
		return nil, yerror.NoSqlDbType
	}
}
