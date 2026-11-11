package sqlstate

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
)

type referenceLeaf struct {
	path digest
	hash digest
}

// referenceRoot builds a canonical compressed tree from a complete set, independently of updates.
func referenceRoot(leaves []referenceLeaf, depth int) digest {
	if len(leaves) == 0 {
		return sha256.Sum256([]byte{0})
	}
	if len(leaves) == 1 {
		return leaves[0].hash
	}
	var left, right []referenceLeaf
	for _, leaf := range leaves {
		if (leaf.path[depth/8]>>uint(7-depth%8))&1 == 0 {
			left = append(left, leaf)
		} else {
			right = append(right, leaf)
		}
	}
	a, b := referenceRoot(left, depth+1), referenceRoot(right, depth+1)
	data := append([]byte{2}, a[:]...)
	return sha256.Sum256(append(data, b[:]...))
}

// TestIncrementalRootsMatchFullReference checks randomized inserts, replacements and deletions.
func TestIncrementalRootsMatchFullReference(t *testing.T) {
	s, table, _ := testStore(t)
	random := rand.New(rand.NewSource(42))
	rows := make(map[string]testOrder)
	parent := ""
	for height := 0; height < 80; height++ {
		id := fmt.Sprintf("block-%d", height)
		must(t, s.Begin(id, parent, uint64(height)))
		key := fmt.Sprintf("key-%d", random.Intn(16))
		if _, exists := rows[key]; exists && random.Intn(3) == 0 {
			must(t, table.Delete(key))
			delete(rows, key)
		} else {
			row := testOrder{ID: key, Status: "open", Amount: uint64(random.Intn(100))}
			must(t, table.Put(row))
			rows[key] = row
		}
		actual := commitTest(t, s)
		var leaves []referenceLeaf
		for _, row := range rows {
			key, payload, err := table.encodeRow(row)
			must(t, err)
			encoded, err := json.Marshal(rowData{Table: table.Name(), Key: key, Data: payload})
			must(t, err)
			rowHash := sha256.Sum256(append([]byte{3}, encoded...))
			path := pathFor(table.Name(), key)
			data := append([]byte{1}, path[:]...)
			leaves = append(leaves, referenceLeaf{path: path, hash: sha256.Sum256(append(data, rowHash[:]...))})
		}
		expected := referenceRoot(leaves, 0)
		if !bytes.Equal(actual, expected[:]) {
			t.Fatalf("root diverged from complete rebuild at height %d", height)
		}
		parent = id
	}
	must(t, s.Finalize(parent))
	var got []testOrder
	must(t, table.Scan(&got))
	if len(got) != len(rows) {
		t.Fatal("GC lost reachable rows")
	}
}
