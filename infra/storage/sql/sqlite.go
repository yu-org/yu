package sql

import (
	"github.com/yu-org/yu/infra/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Sqlite struct {
	*gorm.DB
}

func NewSqlite(dsn string) (*Sqlite, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "/r/n", log.LstdFlags), logger.Config{SlowThreshold: time.Second})
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:          newLogger,
		CreateBatchSize: 50000,
	})
	if err != nil {
		return nil, err
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
