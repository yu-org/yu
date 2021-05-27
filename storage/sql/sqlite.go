package sql

import (
	"github.com/Lawliet-Chan/yu/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Sqlite struct {
	*gorm.DB
}

func NewSqlite(dsn string) (*Sqlite, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

func (*Sqlite) Type() storage.StoreType {
	return storage.Embedded
}

func (*Sqlite) Kind() storage.StoreKind {
	return storage.SQL
}
