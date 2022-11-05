package codec

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/rlp"
)

var GlobalCodec Codec

type Codec interface {
	EncodeToBytes(val any) ([]byte, error)
	DecodeBytes(data []byte, val interface{}) error
}

type RlpCodec struct{}

func (*RlpCodec) EncodeToBytes(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func (*RlpCodec) DecodeBytes(data []byte, val interface{}) error {
	return rlp.DecodeBytes(data, val)
}

type GobCodec struct{}

func (*GobCodec) EncodeToBytes(val interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(val)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (*GobCodec) DecodeBytes(data []byte, val interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(val)
}
