package sql

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"yu/storage"
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

func (*Sqlite) Type() storage.StoreType {
	return storage.Embedded
}

func (*Sqlite) Kind() storage.StoreKind {
	return storage.SQL
}
