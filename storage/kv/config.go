package kv

type KVconf struct {
	// "bolt" "badger" "tikv"
	KVtype string
	// embedded kvdb, such as boltdb, pebble
	Path string
	// distributed kvdb
	Hosts []string
}
