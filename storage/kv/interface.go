package kv

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
}

func NewKV(engine string, fpath string) (KV, error) {
	switch engine {
	case "pebble":
		return NewPebble(fpath)
	case "bolt":
		return NewBolt(fpath)

	default:
		return NewPebble(fpath)
	}
}
