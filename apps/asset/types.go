package asset

import (
	"encoding/binary"
)

type Amount uint64

func (a Amount) Encode() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(a))
	return b
}

func DecodeToAmount(data []byte) Amount {
	return Amount(binary.LittleEndian.Uint64(data))
}
