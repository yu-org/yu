package sql

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/infra/storage"
)

type PostgreSql struct {
	*gorm.DB
}

func NewPostgreSql(cfg *config.SqlDbConf) (*PostgreSql, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "/r/n", log.LstdFlags), logger.Config{SlowThreshold: time.Second})
	db, err := gorm.Open(postgres.Open(cfg.Dsn), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true,
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
