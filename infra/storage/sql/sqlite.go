package sql

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage"
)

type Sqlite struct {
	*gorm.DB
}

func NewSqlite(cfg *config.SqlDbConf) (*Sqlite, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "/r/n", log.LstdFlags), logger.Config{SlowThreshold: time.Second})
	db, err := gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{
		Logger:          newLogger,
		PrepareStmt:     false,
		CreateBatchSize: 50000,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConnections > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	}
	return &Sqlite{db}, nil
}

func (s *Sqlite) Db() *gorm.DB {
	return s.DB
}

func (s *Sqlite) CreateIfNotExist(table interface{}) error {
	if s.Migrator().HasTable(table) {
		return nil
	}
	return s.Migrator().CreateTable(table)
}

func (s *Sqlite) AutoMigrate(table any) error {
	return s.DB.AutoMigrate(table)
}

func (*Sqlite) Type() storage.StoreType {
	return storage.Embedded
}

func (*Sqlite) Kind() storage.StoreKind {
	return storage.SQL
}
