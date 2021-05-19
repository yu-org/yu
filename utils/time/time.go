package time

import "time"

func NowNanoTS() int64 {
	return time.Now().UnixNano()
}

func NowTS() int64 {
	return time.Now().Unix()
}

func NowNanoTsU64() uint64 {
	return uint64(NowNanoTS())
}

func NowTsU64() uint64 {
	return uint64(NowTS())
}
