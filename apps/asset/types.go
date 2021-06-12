package asset

import (
	"github.com/HyperService-Consortium/go-rlp"
	"github.com/sirupsen/logrus"
)

type Amount uint64

func (a Amount) MustEncode() []byte {
	byt, err := a.Encode()
	if err != nil {
		logrus.Panic("encode amount error")
	}
	return byt
}

func (a Amount) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(a)
}

func MustDecodeToAmount(data []byte) Amount {
	a, err := DecodeToAmount(data)
	if err != nil {
		logrus.Panic("decode amount error")
	}
	return a
}

func DecodeToAmount(data []byte) (a Amount, err error) {
	err = rlp.DecodeBytes(data, &a)
	return
}
