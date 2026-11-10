package sqlstate

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"

	"gorm.io/gorm"
)

var (
	ErrNotFound           = errors.New("state row not found")
	ErrVersionUnavailable = errors.New("state version unavailable")
	ErrActiveExecution    = errors.New("finish or abort the active execution first")
)

type versionRecord struct {
	ID     string `gorm:"primaryKey;size:128"`
	Parent string `gorm:"size:128;index"`
	Height uint64
	Root   string `gorm:"size:64"`
}

// TableName identifies retained state versions.
func (versionRecord) TableName() string { return "_yu_state_versions" }

type headRecord struct {
	ID      uint8  `gorm:"primaryKey;autoIncrement:false"`
	Version string `gorm:"size:128"`
}

// TableName identifies the singleton finalized state pointer.
func (headRecord) TableName() string { return "_yu_state_head" }

type pendingChange struct {
	Version string `gorm:"primaryKey;size:128"`
	Path    string `gorm:"primaryKey;size:64"`
	Table   string `gorm:"size:63"`
	Key     []byte
	Data    []byte
	Deleted bool
}

// TableName identifies per-block final row changes, not transaction attempts.
func (pendingChange) TableName() string { return "_yu_state_pending" }

type execution struct {
	version versionRecord
	root    digest
	frames  []map[string]pendingChange
}

// Store owns a single serialized execution stream and any number of pending forks.
// SQL handles may be SQLite, MySQL or PostgreSQL GORM connections. Finalized
// business tables are projections; the tree is the authenticated source of truth.
type Store struct {
	mu       sync.Mutex
	db       *gorm.DB
	tables   map[string]*Table
	active   *execution
	selected string
}

// Open initializes internal tables. A database must not be shared by execution owners.
func Open(db *gorm.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("nil state database")
	}
	migration := db
	if db.Dialector.Name() == "mysql" {
		migration = migration.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_bin")
	}
	if err := migration.AutoMigrate(&storedNode{}, &storedRow{}, &tableRecord{}, &versionRecord{}, &headRecord{}, &pendingChange{}); err != nil {
		return nil, err
	}
	var head headRecord
	err := db.First(&head, "id = ?", 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = db.Create(&headRecord{ID: 1}).Error
	}
	if err != nil {
		return nil, err
	}
	return &Store{db: db, tables: make(map[string]*Table), selected: head.Version}, nil
}

// Begin selects a known parent. Only an empty database accepts an empty parent;
// that first version is height zero and may contain genesis allocations.
func (s *Store) Begin(id, parent string, height uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return ErrActiveExecution
	}
	if id == "" || len(id) > 128 || id == parent {
		return errors.New("invalid state version identity")
	}
	var exists int64
	if err := s.db.Model(&versionRecord{}).Where("id = ?", id).Count(&exists).Error; err != nil {
		return err
	}
	if exists != 0 {
		return errors.New("state version already exists")
	}
	root := emptyRoots[0]
	if parent == "" {
		var count int64
		if err := s.db.Model(&versionRecord{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 0 || height != 0 {
			return ErrVersionUnavailable
		}
	} else {
		prior, err := readVersion(s.db, parent)
		if err != nil {
			return err
		}
		if height == 0 || prior.Height != height-1 {
			return errors.New("state version height does not follow parent")
		}
		root, err = parseDigest(prior.Root)
		if err != nil {
			return err
		}
	}
	s.active = &execution{version: versionRecord{ID: id, Parent: parent, Height: height}, root: root, frames: []map[string]pendingChange{{}}}
	return nil
}

// NextTxn seals a successful transaction and starts an empty transaction frame.
func (s *Store) NextTxn() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return errors.New("no active state execution")
	}
	s.active.frames = append(s.active.frames, make(map[string]pendingChange))
	return nil
}

// DiscardTxn rolls back only the active transaction, retaining earlier successes.
func (s *Store) DiscardTxn() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return errors.New("no active state execution")
	}
	s.active.frames[len(s.active.frames)-1] = make(map[string]pendingChange)
	return nil
}

// Abort discards the whole uncommitted execution without changing persisted roots.
func (s *Store) Abort() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = nil
}

// stage adds a complete immutable copy to the current transaction frame.
func (s *Store) stage(row rowData, deleted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return errors.New("no active state execution")
	}
	path := hexDigest(pathFor(row.Table, row.Key))
	s.active.frames[len(s.active.frames)-1][path] = pendingChange{Path: path, Table: row.Table, Key: bytes.Clone(row.Key), Data: bytes.Clone(row.Data), Deleted: deleted}
	return nil
}

// delete stages a tombstone and keeps the prior row for finalized SQL deletion.
func (s *Store) delete(table string, key []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return errors.New("no active state execution")
	}
	data, err := s.readLocked("", table, key, false)
	if err != nil {
		return err
	}
	path := hexDigest(pathFor(table, key))
	s.active.frames[len(s.active.frames)-1][path] = pendingChange{Path: path, Table: table, Key: bytes.Clone(key), Data: data, Deleted: true}
	return nil
}

// readVersion rejects missing or retired finalized versions explicitly.
func readVersion(db *gorm.DB, id string) (versionRecord, error) {
	var version versionRecord
	err := db.First(&version, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrVersionUnavailable
	}
	return version, err
}

// read synchronizes an authenticated point lookup.
func (s *Store) read(version, table string, key []byte, finalized bool) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(version, table, key, finalized)
}

// readLocked resolves finalized, pending and transaction views; the lock must be held.
func (s *Store) readLocked(version, table string, key []byte, finalized bool) ([]byte, error) {
	root := emptyRoots[0]
	if finalized {
		var head headRecord
		if err := s.db.First(&head, "id = ?", 1).Error; err != nil {
			return nil, err
		}
		version = head.Version
	} else if version == "" && s.active != nil {
		path := hexDigest(pathFor(table, key))
		for i := len(s.active.frames) - 1; i >= 0; i-- {
			if change, ok := s.active.frames[i][path]; ok {
				if change.Deleted {
					return nil, ErrNotFound
				}
				return bytes.Clone(change.Data), nil
			}
		}
		return treeGet(s.db, s.active.root, table, key)
	} else if version == "" {
		version = s.selected
	}
	if version != "" {
		record, err := readVersion(s.db, version)
		if err != nil {
			return nil, err
		}
		root, err = parseDigest(record.Root)
		if err != nil {
			return nil, err
		}
	}
	return treeGet(s.db, root, table, key)
}

// Commit persists a pending version and returns the full-state root.
func (s *Store) Commit() ([]byte, error) { return s.CommitExpected(nil) }

// CommitExpected checks a received block's root inside the SQL transaction.
// A mismatch rolls back nodes, rows, changes and version metadata together.
func (s *Store) CommitExpected(expected []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return nil, errors.New("no active state execution")
	}
	combined := make(map[string]pendingChange)
	for _, frame := range s.active.frames {
		for key, change := range frame {
			combined[key] = change
		}
	}
	keys := make([]string, 0, len(combined))
	for key := range combined {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	root := s.active.root
	version := s.active.version
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			change := combined[key]
			var err error
			root, err = treeUpdate(tx, root, rowData{Table: change.Table, Key: change.Key, Data: change.Data}, change.Deleted)
			if err != nil {
				return err
			}
			change.Version = version.ID
			if err := tx.Create(&change).Error; err != nil {
				return err
			}
		}
		if expected != nil && !bytes.Equal(expected, root[:]) {
			return errors.New("state root mismatch")
		}
		version.Root = hexDigest(root)
		return tx.Create(&version).Error
	})
	if err != nil {
		return nil, err
	}
	s.selected, s.active = version.ID, nil
	return bytes.Clone(root[:]), nil
}

// Root returns a retained pending root or the latest finalized root.
func (s *Store) Root(version string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := readVersion(s.db, version)
	if err != nil {
		return nil, err
	}
	root, err := parseDigest(record.Root)
	return bytes.Clone(root[:]), err
}

// Finalized returns the latest finalized version identity, or an empty string before genesis.
func (s *Store) Finalized() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var head headRecord
	err := s.db.First(&head, "id = ?", 1).Error
	return head.Version, err
}

// Finalize promotes an executed descendant atomically across registered tables.
// Only the new finalized version and its pending descendants survive.
func (s *Store) Finalize(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" {
		return ErrVersionUnavailable
	}
	if s.active != nil {
		return ErrActiveExecution
	}
	kept := make(map[string]bool)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var head headRecord
		if err := tx.First(&head, "id = ?", 1).Error; err != nil {
			return err
		}
		if id == head.Version {
			kept[s.selected] = true
			return nil
		}
		var path []versionRecord
		cursor := id
		for cursor != head.Version {
			if cursor == "" {
				return errors.New("cannot finalize a fork outside the finalized ancestry")
			}
			version, err := readVersion(tx, cursor)
			if err != nil {
				return err
			}
			path = append(path, version)
			cursor = version.Parent
		}
		for i := len(path) - 1; i >= 0; i-- {
			var changes []pendingChange
			if err := tx.Where("version = ?", path[i].ID).Order("path").Find(&changes).Error; err != nil {
				return err
			}
			for _, change := range changes {
				table := s.tables[change.Table]
				if table == nil {
					return fmt.Errorf("register table %s before finalizing", change.Table)
				}
				if err := table.materialize(tx, change); err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&headRecord{}).Where("id = ?", 1).Update("version", id).Error; err != nil {
			return err
		}
		var versions []versionRecord
		if err := tx.Find(&versions).Error; err != nil {
			return err
		}
		kept[id] = true
		// Repeated passes handle arbitrary SQL row order without relying on height.
		for changed := true; changed; {
			changed = false
			for _, version := range versions {
				if !kept[version.ID] && kept[version.Parent] {
					kept[version.ID], changed = true, true
				}
			}
		}
		for _, version := range versions {
			if !kept[version.ID] {
				if err := removeVersion(tx, version.ID); err != nil {
					return err
				}
			}
		}
		if err := tx.Where("version = ?", id).Delete(&pendingChange{}).Error; err != nil {
			return err
		}
		return collectGarbage(tx)
	})
	if err == nil && !kept[s.selected] {
		s.selected = id
	}
	return err
}

// Bind renames a local execution version to its sealed block hash without changing
// state content. Call it after computing StateRoot and sealing the block.
func (s *Store) Bind(localID, blockHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return ErrActiveExecution
	}
	if blockHash == "" || len(blockHash) > 128 {
		return errors.New("invalid sealed block identity")
	}
	if _, err := readVersion(s.db, localID); err != nil {
		return err
	}
	if localID == blockHash {
		return nil
	}
	var head headRecord
	if err := s.db.First(&head, "id = ?", 1).Error; err != nil {
		return err
	}
	if head.Version == localID {
		return errors.New("cannot rename finalized state")
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&versionRecord{}).Where("id = ?", localID).Update("id", blockHash).Error; err != nil {
			return err
		}
		if err := tx.Model(&versionRecord{}).Where("parent = ?", localID).Update("parent", blockHash).Error; err != nil {
			return err
		}
		if err := tx.Model(&pendingChange{}).Where("version = ?", localID).Update("version", blockHash).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil && s.selected == localID {
		s.selected = blockHash
	}
	return err
}

// removeVersion removes its metadata and pending changes; the caller owns a transaction.
func removeVersion(db *gorm.DB, id string) error {
	if err := db.Where("version = ?", id).Delete(&pendingChange{}).Error; err != nil {
		return err
	}
	return db.Where("id = ?", id).Delete(&versionRecord{}).Error
}

// DropPending discards a pending block and its descendants, retaining the finalized version.
func (s *Store) DropPending(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return ErrActiveExecution
	}
	dropped := map[string]bool{id: true}
	var finalized string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var head headRecord
		if err := tx.First(&head, "id = ?", 1).Error; err != nil {
			return err
		}
		finalized = head.Version
		if id == finalized {
			return errors.New("cannot drop finalized state")
		}
		if _, err := readVersion(tx, id); err != nil {
			return err
		}
		var versions []versionRecord
		if err := tx.Find(&versions).Error; err != nil {
			return err
		}
		for changed := true; changed; {
			changed = false
			for _, version := range versions {
				if !dropped[version.ID] && dropped[version.Parent] {
					dropped[version.ID], changed = true, true
				}
			}
		}
		for version := range dropped {
			if err := removeVersion(tx, version); err != nil {
				return err
			}
		}
		return collectGarbage(tx)
	})
	if err == nil && dropped[s.selected] {
		s.selected = finalized
	}
	return err
}

// collectGarbage retains exactly the nodes and values reachable from retained roots.
func collectGarbage(db *gorm.DB) error {
	nodes, rows := make(map[string]bool), make(map[string]bool)
	var versions []versionRecord
	if err := db.Find(&versions).Error; err != nil {
		return err
	}
	for _, version := range versions {
		root, err := parseDigest(version.Root)
		if err != nil {
			return err
		}
		if err := walkTree(db, root, 0, nodes, rows, nil); err != nil {
			return err
		}
	}
	if err := sweep(db, &storedNode{}, nodes); err != nil {
		return err
	}
	return sweep(db, &storedRow{}, rows)
}

// sweep deletes unreachable content in bounded batches supported by all SQL drivers.
func sweep(db *gorm.DB, model any, retained map[string]bool) error {
	var hashes []string
	if err := db.Model(model).Pluck("hash", &hashes).Error; err != nil {
		return err
	}
	var batch []string
	for _, hash := range hashes {
		if !retained[hash] {
			batch = append(batch, hash)
		}
		if len(batch) == 256 {
			if err := db.Where("hash IN ?", batch).Delete(model).Error; err != nil {
				return err
			}
			batch = nil
		}
	}
	if len(batch) != 0 {
		return db.Where("hash IN ?", batch).Delete(model).Error
	}
	return nil
}
