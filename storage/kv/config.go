package kv

type KVconf struct {
	// embedded kvdb, such as boltdb, pebble
	path string
	// distributed kvdb
	hosts []string
}
