package sqlstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

var identifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type tableRecord struct {
	Name     string `gorm:"primaryKey;size:63"`
	Tripod   string `gorm:"size:63"`
	Business string `gorm:"size:63"`
	Schema   string `gorm:"type:text"`
}

// TableName identifies persisted, immutable table definitions.
func (tableRecord) TableName() string { return "_yu_state_tables" }

// Table describes one concrete row type. Writes must go through this handle;
// direct SQL writes bypass state commitments and are not supported.
type Table struct {
	store  *Store
	name   string
	typ    reflect.Type
	fields []*schema.Field
	keys   []*schema.Field
	record tableRecord
}

type encodedField struct {
	Column string
	Value  json.RawMessage
}

// RegisterTable derives the business name using GORM, honoring a static TableName
// method or the configured naming strategy, and prefixes it with the tripod name.
// It creates missing SQL tables and reuses matching registrations idempotently.
// Names must be lowercase SQL identifiers and produce a name of at most 63 bytes.
// Auto-generated fields, associations and custom GORM serializers are rejected;
// all consensus values must be supplied explicitly by the caller.
func (s *Store) RegisterTable(tripod string, model any) (*Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return nil, errors.New("register tables before beginning execution")
	}
	typ := reflect.TypeOf(model)
	if typ == nil {
		return nil, errors.New("state model is nil")
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("state model must be a struct")
	}
	parsed, err := schema.Parse(reflect.New(typ).Interface(), &sync.Map{}, s.db.NamingStrategy)
	if err != nil {
		return nil, err
	}
	business := parsed.Table
	name := tripod + "_" + business
	if !identifier.MatchString(tripod) || !identifier.MatchString(business) || len(name) > 63 {
		return nil, errors.New("invalid state table name")
	}
	if len(parsed.PrimaryFields) == 0 || len(parsed.Relationships.Relations) != 0 {
		return nil, errors.New("state models require primary keys and cannot contain associations")
	}
	for _, index := range parsed.ParseIndexes() {
		if index.Class == "UNIQUE" {
			return nil, errors.New("secondary uniqueness must be checked against pending state")
		}
	}
	table := &Table{store: s, name: name, typ: typ, keys: parsed.PrimaryFields}
	var signature []string
	for _, field := range parsed.Fields {
		if field.DBName == "" {
			continue
		}
		if field.AutoIncrement || field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 || field.HasDefaultValue || field.Serializer != nil {
			return nil, fmt.Errorf("column %s must not use generated values, defaults or serializers", field.DBName)
		}
		if !field.Readable || !field.Creatable || !field.Updatable {
			return nil, fmt.Errorf("column %s must support reads and writes", field.DBName)
		}
		if field.Unique || field.TagSettings["UNIQUEINDEX"] != "" || field.TagSettings["CHECK"] != "" {
			return nil, fmt.Errorf("column %s requires a constraint not enforced by pending views", field.DBName)
		}
		fieldType := field.FieldType
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		kind := fieldType.Kind()
		if kind != reflect.String && kind != reflect.Bool && (kind < reflect.Int || kind > reflect.Uint64) && !(kind == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8) {
			return nil, fmt.Errorf("column %s must be a portable scalar or byte slice", field.DBName)
		}
		if field.TagSettings["TYPE"] != "" {
			return nil, fmt.Errorf("column %s cannot override the SQL type", field.DBName)
		}
		if field.PrimaryKey {
			kind := field.FieldType.Kind()
			if kind != reflect.String && (kind < reflect.Int || kind > reflect.Uint64) {
				return nil, fmt.Errorf("primary key %s must be a string or integer", field.DBName)
			}
		}
		table.fields = append(table.fields, field)
	}
	sort.Slice(table.fields, func(i, j int) bool { return table.fields[i].DBName < table.fields[j].DBName })
	sort.Slice(table.keys, func(i, j int) bool { return table.keys[i].DBName < table.keys[j].DBName })
	for _, field := range table.fields {
		signature = append(signature, fmt.Sprintf("%s:%s:%s:%t:%s", field.DBName, field.FieldType.PkgPath(), field.FieldType, field.PrimaryKey, field.Tag))
	}
	encoded, _ := json.Marshal(signature)
	table.record = tableRecord{Name: name, Tripod: tripod, Business: business, Schema: string(encoded)}
	if prior := s.tables[name]; prior != nil {
		if prior.record != table.record || prior.typ != typ {
			return nil, errors.New("conflicting state table registration")
		}
		return prior, nil
	}
	var prior tableRecord
	err = s.db.First(&prior, "name = ?", name).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil && prior != table.record {
		return nil, errors.New("state table schema changed; explicit migration required")
	}
	// Migration is setup work: some SQL drivers implicitly commit DDL.
	migration := s.db.Table(name)
	if s.db.Dialector.Name() == "mysql" {
		migration = migration.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_bin")
	}
	if err := migration.AutoMigrate(reflect.New(typ).Interface()); err != nil {
		return nil, err
	}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&table.record).Error; err != nil {
		return nil, err
	}
	s.tables[name] = table
	return table, nil
}

// Name returns the physical SQL table name, including the tripod namespace.
func (t *Table) Name() string { return t.name }

// rowValue validates and dereferences a model instance.
func (t *Table) rowValue(row any) (reflect.Value, error) {
	value := reflect.ValueOf(row)
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, errors.New("nil state row")
		}
		value = value.Elem()
	}
	if !value.IsValid() || value.Type() != t.typ {
		return reflect.Value{}, fmt.Errorf("expected row type %s", t.typ)
	}
	return value, nil
}

// encodeRow encodes only persisted fields, in column order, including primary keys.
func (t *Table) encodeRow(row any) ([]byte, []byte, error) {
	value, err := t.rowValue(row)
	if err != nil {
		return nil, nil, err
	}
	fields := make([]encodedField, 0, len(t.fields))
	for _, field := range t.fields {
		v, _ := field.ValueOf(context.Background(), value)
		if err := validateSQLValue(field, v); err != nil {
			return nil, nil, err
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, nil, fmt.Errorf("encoding %s: %w", field.DBName, err)
		}
		fields = append(fields, encodedField{Column: field.DBName, Value: encoded})
	}
	key, err := t.keyFromRow(value)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(fields)
	return key, data, err
}

// validateSQLValue rejects values that cannot retain their identity across SQL drivers.
func validateSQLValue(field *schema.Field, value any) error {
	if value == nil {
		if field.NotNull || field.PrimaryKey {
			return fmt.Errorf("column %s cannot be null", field.DBName)
		}
		return nil
	}
	v := reflect.ValueOf(value)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if field.NotNull {
				return fmt.Errorf("column %s cannot be null", field.DBName)
			}
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice && v.IsNil() && field.NotNull {
		return fmt.Errorf("column %s cannot be null", field.DBName)
	}
	if v.Kind() == reflect.String && (!utf8.ValidString(v.String()) || (field.PrimaryKey && strings.HasSuffix(v.String(), " "))) {
		return fmt.Errorf("column %s has a nonportable string identity", field.DBName)
	}
	if v.Kind() == reflect.String {
		if field.Size > 0 && utf8.RuneCountInString(v.String()) > field.Size {
			return fmt.Errorf("column %s exceeds its declared size", field.DBName)
		}
		if (field.PrimaryKey || field.TagSettings["INDEX"] != "") && len(v.String()) > 191 {
			return fmt.Errorf("indexed column %s exceeds portable index length", field.DBName)
		}
	}
	if v.Kind() >= reflect.Uint && v.Kind() <= reflect.Uint64 && v.Uint() > math.MaxInt64 {
		return fmt.Errorf("column %s exceeds portable SQL integer range; use a decimal string", field.DBName)
	}
	return nil
}

// keyFromRow encodes primary keys in their registered column order.
func (t *Table) keyFromRow(row reflect.Value) ([]byte, error) {
	var key []any
	for _, field := range t.keys {
		v, _ := field.ValueOf(context.Background(), row)
		key = append(key, v)
	}
	return json.Marshal(key)
}

// encodeKey normalizes a scalar primary key or a model holding composite keys.
func (t *Table) encodeKey(key any) ([]byte, error) {
	if value, err := t.rowValue(key); err == nil {
		return t.keyFromRow(value)
	}
	if len(t.keys) != 1 {
		return nil, errors.New("composite keys require a model holding its primary key fields")
	}
	encoded, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}
	value := reflect.New(t.keys[0].FieldType)
	if err := json.Unmarshal(encoded, value.Interface()); err != nil {
		return nil, fmt.Errorf("invalid primary key: %w", err)
	}
	return json.Marshal([]any{value.Elem().Interface()})
}

// decodeRow restores the registered struct and rejects inconsistent persisted data.
func (t *Table) decodeRow(data []byte) (reflect.Value, error) {
	var fields []encodedField
	if err := json.Unmarshal(data, &fields); err != nil {
		return reflect.Value{}, err
	}
	if len(fields) != len(t.fields) {
		return reflect.Value{}, errors.New("state row schema mismatch")
	}
	row := reflect.New(t.typ)
	for i, field := range t.fields {
		if fields[i].Column != field.DBName {
			return reflect.Value{}, errors.New("state row column mismatch")
		}
		value := reflect.New(field.FieldType)
		if err := json.Unmarshal(fields[i].Value, value.Interface()); err != nil {
			return reflect.Value{}, err
		}
		if err := field.Set(context.Background(), row.Elem(), value.Elem().Interface()); err != nil {
			return reflect.Value{}, err
		}
	}
	return row, nil
}

// Put stages a complete row in the current transaction; primary keys are immutable identities.
func (t *Table) Put(row any) error {
	key, data, err := t.encodeRow(row)
	if err != nil {
		return err
	}
	return t.store.stage(rowData{Table: t.name, Key: key, Data: data}, false)
}

// Delete stages a deletion, retaining the old row for SQL materialization.
func (t *Table) Delete(key any) error {
	encoded, err := t.encodeKey(key)
	if err != nil {
		return err
	}
	return t.store.delete(t.name, encoded)
}

// Get reads the selected execution/pending view into a pointer to the registered row type.
func (t *Table) Get(key, into any) error { return t.read("", key, into, false) }

// GetFinalized reads only the latest finalized view.
func (t *Table) GetFinalized(key, into any) error { return t.read("", key, into, true) }

// GetPending reads a retained pending version or the latest finalized version.
// Older finalized versions are unavailable after finalization.
func (t *Table) GetPending(version string, key, into any) error {
	return t.read(version, key, into, false)
}

// Scan reads the complete selected table view in deterministic encoded-key order.
// Pending scans merge transaction frames over the parent tree; SQL secondary
// indexes are used by FindFinalized only, never as a stale pending-state filter.
func (t *Table) Scan(into any) error { return t.ScanPending("", into) }

// ScanPending reads a retained pending version, or the current view when empty.
// The destination must be *[]Row or *[]*Row for the registered model.
func (t *Table) ScanPending(version string, into any) error {
	destination := reflect.ValueOf(into)
	if destination.Kind() != reflect.Pointer || destination.IsNil() || destination.Elem().Kind() != reflect.Slice {
		return errors.New("scan destination must be a pointer to a row slice")
	}
	element := destination.Elem().Type().Elem()
	if element != t.typ && element != reflect.PointerTo(t.typ) {
		return errors.New("scan row type does not match table")
	}
	s := t.store
	s.mu.Lock()
	defer s.mu.Unlock()
	root := emptyRoots[0]
	frames := []map[string]pendingChange(nil)
	if version == "" && s.active != nil {
		root, frames = s.active.root, s.active.frames
	} else {
		if version == "" {
			version = s.selected
		}
		if version != "" {
			record, err := readVersion(s.db, version)
			if err != nil {
				return err
			}
			root, err = parseDigest(record.Root)
			if err != nil {
				return err
			}
		}
	}
	rows := make(map[string][]byte)
	if err := walkTree(s.db, root, 0, make(map[string]bool), make(map[string]bool), func(row rowData) error {
		if row.Table == t.name {
			rows[string(row.Key)] = row.Data
		}
		return nil
	}); err != nil {
		return err
	}
	for _, frame := range frames {
		for _, change := range frame {
			if change.Table != t.name {
				continue
			}
			if change.Deleted {
				delete(rows, string(change.Key))
			} else {
				rows[string(change.Key)] = change.Data
			}
		}
	}
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := reflect.MakeSlice(destination.Elem().Type(), 0, len(keys))
	for _, key := range keys {
		row, err := t.decodeRow(rows[key])
		if err != nil {
			return err
		}
		if element == t.typ {
			row = row.Elem()
		}
		result = reflect.Append(result, row)
	}
	destination.Elem().Set(result)
	return nil
}

// read validates a destination and decodes a row from the requested view.
func (t *Table) read(version string, key, into any, finalized bool) error {
	destination := reflect.ValueOf(into)
	if destination.Kind() != reflect.Pointer || destination.IsNil() || destination.Elem().Type() != t.typ {
		return fmt.Errorf("destination must be *%s", t.typ)
	}
	encoded, err := t.encodeKey(key)
	if err != nil {
		return err
	}
	data, err := t.store.read(version, t.name, encoded, finalized)
	if err != nil {
		return err
	}
	row, err := t.decodeRow(data)
	if err == nil {
		destination.Elem().Set(row.Elem())
	}
	return err
}

// materialize applies one committed change without hooks or generated values.
func (t *Table) materialize(db *gorm.DB, change pendingChange) error {
	row, err := t.decodeRow(change.Data)
	if err != nil {
		return err
	}
	values := make(map[string]any, len(t.fields))
	for _, field := range t.fields {
		value, _ := field.ValueOf(context.Background(), row.Elem())
		values[field.DBName] = value
	}
	query := db.Session(&gorm.Session{SkipHooks: true}).Table(t.name)
	if change.Deleted {
		keys := make(map[string]any, len(t.keys))
		for _, field := range t.keys {
			keys[field.DBName] = values[field.DBName]
		}
		return query.Where(keys).Delete(reflect.New(t.typ).Interface()).Error
	}
	var keys []clause.Column
	var columns []string
	for _, field := range t.keys {
		keys = append(keys, clause.Column{Name: field.DBName})
	}
	for _, field := range t.fields {
		columns = append(columns, field.DBName)
	}
	return query.Clauses(clause.OnConflict{Columns: keys, DoUpdates: clause.AssignmentColumns(columns)}).Create(values).Error
}

// FindFinalized queries real SQL columns of the latest finalized table.
// Conditions must name registered columns; results are ordered by primary key.
func (t *Table) FindFinalized(equal map[string]any, into any) error {
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	query := t.store.db.Table(t.name)
	for column := range equal {
		found := false
		for _, field := range t.fields {
			found = found || field.DBName == column
		}
		if !found {
			return fmt.Errorf("unknown column %s", column)
		}
	}
	query = query.Where(equal)
	for _, field := range t.keys {
		query = query.Order(clause.OrderByColumn{Column: clause.Column{Name: field.DBName}})
	}
	return query.Find(into).Error
}
