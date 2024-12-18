package sql

import (
	"github.com/yu-org/yu/infra/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"time"
)

type PostgreSql struct {
	*gorm.DB
}

func NewPostgreSql(dsn string) (*PostgreSql, error) {
	newLogger := logger.New(os.Stdout, logger.Config{SlowThreshold: time.Second})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	return &PostgreSql{db}, nil
}

func (p *PostgreSql) Db() *gorm.DB {
	return p.DB
}

func (p *PostgreSql) CreateIfNotExist(table interface{}) error {
	if p.Migrator().HasTable(table) {
		return nil
	}
	return p.Migrator().CreateTable(table)
}

func (p *PostgreSql) AutoMigrate(table any) error {
	return p.DB.AutoMigrate(table)
}

func (*PostgreSql) Type() storage.StoreType {
	return storage.Server
}

func (*PostgreSql) Kind() storage.StoreKind {
	return storage.SQL
}
