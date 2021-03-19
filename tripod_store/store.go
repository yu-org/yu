package tripod_store

type TripodStore struct {
	kvdb *TripodKV
}

func (ts *TripodStore) Commit() {
	ts.kvdb.Commit()
}

func (ts *TripodStore) Rollback() {
	ts.kvdb.Rollback()
}

func (ts *TripodStore) Flush() {
	ts.kvdb.Flush()
}
