package sql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"yu/storage"
)

type Mysql struct {
	*gorm.DB
}

func NewMysql(dsn string) (*Mysql, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Mysql{db}, nil
}

func (m *Mysql) Db() *gorm.DB {
	return m.DB
}

func (m *Mysql) CreateIfNotExist(table interface{}) error {
	if m.Migrator().HasTable(table) {
		return nil
	}
	return m.Migrator().CreateTable(table)
}

func (*Mysql) Type() storage.StoreType {
	return storage.Server
}

func (*Mysql) Kind() storage.StoreKind {
	return storage.SQL
}
