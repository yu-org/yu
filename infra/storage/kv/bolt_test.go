package kv

import (
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
	"os"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	db, err := bbolt.Open("testdb", 0666, nil)
	assert.NoError(t, err)
	start := time.Now()
	err = db.Update(func(tx *bbolt.Tx) error {
		bu, err := tx.CreateBucketIfNotExists([]byte("bu"))
		if err != nil {
			return err
		}
		return bu.Put([]byte("key"), []byte("value"))
	})
	assert.NoError(t, err)
	t.Logf("%d ms", time.Since(start).Milliseconds())
	os.RemoveAll("testdb")
}
