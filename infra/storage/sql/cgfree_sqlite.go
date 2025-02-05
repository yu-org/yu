package sql

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/yu-org/yu/config"
)

var GlobalSqliteManager *CgofreeSqliteManager

func Init(conf config.SqliteDBConf) {
	GlobalSqliteManager = &CgofreeSqliteManager{
		conf: conf,
	}
}

type CgofreeSqliteManager struct {
	sync.Mutex
	conf config.SqliteDBConf
	init bool
}

func (m *CgofreeSqliteManager) initial() error {
	if m.init {
		return nil
	}
	dir := m.conf.Path
	if _, err := os.Stat(m.conf.Path); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	m.init = true
	return nil
}

func (m *CgofreeSqliteManager) CreateStore(name string) (*sql.DB, error) {
	m.Lock()
	defer m.Unlock()
	if err := m.initial(); err != nil {
		return nil, err
	}
	dbPath := path.Join(m.conf.Path, name)
	db, err := sql.Open("sqlite", connectionString(dbPath))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectionString(dpath string) string {
	return fmt.Sprintf("file:%s?cache=shared&_journal=WAL&sync=2", dpath)
}
