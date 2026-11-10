package sqlstate

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/types"
	"gorm.io/gorm"
)

type rawRow struct {
	ID    string `gorm:"primaryKey;size:64"`
	Key   []byte
	Value []byte
}

// TableName preserves the compatibility table suffix across model renames.
func (rawRow) TableName() string { return "kv" }

// State adapts table state to yu's legacy IState lifecycle. New applications
// should register concrete tables; the kv table is only for existing Set/Get users.
// Errors from legacy void methods are exposed through ExecutionError.
type State struct {
	*Store
	mu        sync.Mutex
	raw       map[string]*Table
	block     *types.Block
	parent    common.Hash
	expected  common.Hash
	sealed    common.Hash
	height    common.BlockNum
	version   string
	committed bool
	lastErr   error
}

var _ state.IState = (*State)(nil)
var _ state.ExecutionState = (*State)(nil)

// NewState creates the SQL backend over an already configured SQL driver.
func NewState(db *gorm.DB) (*State, error) {
	store, err := Open(db)
	if err != nil {
		return nil, err
	}
	return &State{Store: store, raw: make(map[string]*Table)}, nil
}

// GenesisStateRoot commits to the empty state used by the default genesis builder.
func (s *State) GenesisStateRoot() common.Hash { return common.BytesToHash(emptyRoots[0][:]) }

// PrepareTripod registers the backwards-compatible KV table before execution.
func (s *State) PrepareTripod(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.raw[name] != nil {
		return nil
	}
	table, err := s.RegisterTable(name, &rawRow{})
	if err == nil {
		s.raw[name] = table
	}
	return err
}

// InitializeState anchors the empty genesis state once. Custom genesis state must
// be initialized through Store before startup, using the actual genesis identity.
func (s *State) InitializeState(genesis *types.Block) error {
	if genesis == nil || genesis.Header == nil || genesis.Hash == common.NullHash {
		return errors.New("invalid state genesis")
	}
	head, err := s.Store.Finalized()
	if err != nil {
		return err
	}
	if head != "" {
		return s.checkGenesis(genesis.Hash.String())
	}
	if err := s.checkGenesis(genesis.Hash.String()); err != nil {
		return err
	}
	id := genesis.Hash.String()
	if _, err := s.Store.Root(id); errors.Is(err, ErrVersionUnavailable) {
		if err := s.Store.Begin(id, "", 0); err != nil {
			return err
		}
		var expected []byte
		if genesis.StateRoot != common.NullHash {
			expected = genesis.StateRoot.Bytes()
		}
		if _, err := s.Store.CommitExpected(expected); err != nil {
			s.Store.Abort()
			return err
		}
	} else if err != nil {
		return err
	}
	return s.Store.Finalize(id)
}

type chainRecord struct {
	ID      uint8  `gorm:"primaryKey;autoIncrement:false"`
	Genesis string `gorm:"size:128"`
}

// TableName identifies the chain binding independent of pruned genesis versions.
func (chainRecord) TableName() string { return "_yu_state_chain" }

// checkGenesis prevents opening a state database for another chain.
func (s *State) checkGenesis(hash string) error {
	if err := s.Store.db.AutoMigrate(&chainRecord{}); err != nil {
		return err
	}
	var record chainRecord
	err := s.Store.db.First(&record, "id = ?", 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.Store.db.Create(&chainRecord{ID: 1, Genesis: hash}).Error
	}
	if err != nil {
		return err
	}
	if record.Genesis != hash {
		return errors.New("state database belongs to another genesis")
	}
	return nil
}

// BeginExecution selects the parent state for both production and history sync.
func (s *State) BeginExecution(block *types.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if block == nil || block.Header == nil {
		return errors.New("invalid execution block")
	}
	if s.block == block && !s.committed {
		return s.lastErr
	}
	if s.block != nil && !s.committed {
		return ErrActiveExecution
	}
	if s.block != nil && s.committed && s.version != s.block.Hash.String() {
		return errors.New("bind or abort the previous execution before beginning another")
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return err
	}
	version := "execution_" + hex.EncodeToString(id[:])
	if err := s.Store.Begin(version, block.PrevHash.String(), uint64(block.Height)); err != nil {
		return err
	}
	s.block, s.parent, s.height = block, block.PrevHash, block.Height
	s.expected, s.sealed = block.StateRoot, block.Hash
	s.version, s.committed, s.lastErr = version, false, nil
	return nil
}

// StartBlock preserves the legacy lifecycle; BeginExecution also returns errors directly.
func (s *State) StartBlock(block *types.Block) {
	if err := s.BeginExecution(block); err != nil {
		s.mu.Lock()
		s.lastErr = err
		s.mu.Unlock()
	}
}

// ExecutionError reports an error from legacy void state operations.
func (s *State) ExecutionError() error { s.mu.Lock(); defer s.mu.Unlock(); return s.lastErr }

// ExecutionCommitted reports whether this block has executed and been bound.
func (s *State) ExecutionCommitted(block *types.Block) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.block == block && s.committed && s.version == block.Hash.String()
}

// Set stages a legacy KV row. PrepareTripod must have run during registration.
func (s *State) Set(tripod state.NameString, key, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastErr != nil {
		return
	}
	table := s.raw[tripod.Name()]
	if table == nil {
		s.lastErr = errors.New("tripod state table is not prepared")
		return
	}
	id := sha256.Sum256(key)
	s.lastErr = table.Put(rawRow{ID: hex.EncodeToString(id[:]), Key: bytes.Clone(key), Value: bytes.Clone(value)})
}

// Get reads a legacy KV value from the selected view, returning nil for absence.
func (s *State) Get(tripod state.NameString, key []byte) ([]byte, error) {
	return s.getRaw(tripod, key, "", false)
}

// GetFinalized reads only the latest finalized KV state.
func (s *State) GetFinalized(tripod state.NameString, key []byte) ([]byte, error) {
	return s.getRaw(tripod, key, "", true)
}

// GetByBlockHash supports retained pending blocks and the latest finalized block only.
func (s *State) GetByBlockHash(tripod state.NameString, key []byte, block *types.Block) ([]byte, error) {
	if block == nil || block.Header == nil {
		return nil, errors.New("invalid block state lookup")
	}
	return s.getRaw(tripod, key, block.Hash.String(), false)
}

// getRaw decodes and verifies the original key in the compatibility table.
func (s *State) getRaw(tripod state.NameString, key []byte, version string, finalized bool) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	table := s.raw[tripod.Name()]
	if table == nil {
		return nil, errors.New("tripod state table is not prepared")
	}
	id := sha256.Sum256(key)
	var row rawRow
	err := table.read(version, hex.EncodeToString(id[:]), &row, finalized)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(row.Key, key) {
		return nil, errors.New("state key hash collision")
	}
	return bytes.Clone(row.Value), nil
}

// Exist reports legacy KV membership, including values that encode an empty slice.
func (s *State) Exist(tripod state.NameString, key []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	table := s.raw[tripod.Name()]
	if table == nil {
		s.lastErr = errors.New("tripod state table is not prepared")
		return false
	}
	id := sha256.Sum256(key)
	var row rawRow
	err := table.Get(hex.EncodeToString(id[:]), &row)
	if errors.Is(err, ErrNotFound) {
		return false
	}
	if err != nil {
		s.lastErr = err
		return false
	}
	return bytes.Equal(row.Key, key)
}

// Delete stages a legacy KV deletion; deleting an absent key is a no-op.
func (s *State) Delete(tripod state.NameString, key []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastErr != nil {
		return
	}
	table := s.raw[tripod.Name()]
	if table == nil {
		s.lastErr = errors.New("tripod state table is not prepared")
		return
	}
	id := sha256.Sum256(key)
	err := table.Delete(hex.EncodeToString(id[:]))
	if !errors.Is(err, ErrNotFound) {
		s.lastErr = err
	}
}

// NextTxn preserves successful writes and opens the next transaction frame.
func (s *State) NextTxn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastErr == nil {
		s.lastErr = s.Store.NextTxn()
	}
}

// Discard drops the failed transaction and clears its legacy write error.
func (s *State) Discard() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastErr = s.Store.DiscardTxn()
}

// AbortExecution drops an active or committed pending execution after failure.
// A version already retired by fork choice is harmless; finalized state is retained.
func (s *State) AbortExecution() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Store.Abort()
	var abortErr error
	if s.committed && s.version != "" {
		head, err := s.Store.Finalized()
		abortErr = err
		if err == nil && head != s.version {
			err = s.Store.DropPending(s.version)
			if !errors.Is(err, ErrVersionUnavailable) {
				abortErr = err
			}
		}
	}
	s.block, s.version, s.committed, s.lastErr = nil, "", false, abortErr
}

// DiscardAll aborts the block rather than rewriting a root index to mutated nodes.
func (s *State) DiscardAll() { s.AbortExecution() }

// Commit atomically verifies a received root and persists the resulting pending state.
func (s *State) Commit() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastErr != nil {
		return nil, s.lastErr
	}
	if s.block == nil || s.committed {
		return nil, errors.New("no active execution")
	}
	var expected []byte
	if s.expected != common.NullHash {
		expected = s.expected.Bytes()
	}
	root, err := s.Store.CommitExpected(expected)
	if err == nil {
		s.committed = true
	}
	return root, err
}

// BindBlock binds a committed execution to the final sealed block identity.
// The caller must bind before appending, finalizing or executing a child block.
func (s *State) BindBlock(block *types.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.committed || s.block != block || block.Hash == common.NullHash || block.PrevHash != s.parent || block.Height != s.height {
		return errors.New("block does not match the committed execution")
	}
	if s.sealed != common.NullHash && block.Hash != s.sealed {
		return errors.New("received block identity changed during execution")
	}
	if err := types.VerifyHeaderCommitments(block); err != nil {
		return err
	}
	root, err := s.Store.Root(s.version)
	if err != nil {
		return err
	}
	if !bytes.Equal(root, block.StateRoot.Bytes()) {
		return errors.New("sealed block state root mismatch")
	}
	if err := s.Store.Bind(s.version, block.Hash.String()); err != nil {
		return err
	}
	s.version = block.Hash.String()
	return nil
}

// FinalizeState promotes only a known, bound block. Errors leave the prior head intact.
func (s *State) FinalizeState(block *types.Block) error {
	if block == nil || block.Header == nil {
		return errors.New("invalid finalized block")
	}
	return s.Store.Finalize(block.Hash.String())
}

// FinalizedStateHead returns only the latest finalized block and its full-state root.
func (s *State) FinalizedStateHead() (common.Hash, common.Hash, error) {
	id, err := s.Store.Finalized()
	if err != nil {
		return common.NullHash, common.NullHash, err
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(id, "0x"))
	if err != nil || len(decoded) != 32 {
		return common.NullHash, common.NullHash, errors.New("finalized state is not bound to a block hash")
	}
	root, err := s.Store.Root(id)
	return common.BytesToHash(decoded), common.BytesToHash(root), err
}

// FinalizeBlock exposes the legacy void callback; new consensus code should use FinalizeState.
func (s *State) FinalizeBlock(block *types.Block) {
	if err := s.FinalizeState(block); err != nil {
		s.mu.Lock()
		s.lastErr = err
		s.mu.Unlock()
	}
}
