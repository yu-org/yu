package storage

type StorageType interface {
	Type() StoreType
	Kind() StoreKind
}

type StoreType int

const (
	Embedded StoreType = iota
	Server
)

type StoreKind int

const (
	KV StoreKind = iota
	SQL
	FS
)
