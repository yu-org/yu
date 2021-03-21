package sql

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"yu/storage"
)

type PostgreSql struct {
	*gorm.DB
}

func NewPostgreSql(dsn string) (*PostgreSql, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &PostgreSql{db}, nil
}

func (*PostgreSql) Type() storage.StoreType {
	return storage.Server
}

func (*PostgreSql) Kind() storage.StoreKind {
	return storage.SQL
}
