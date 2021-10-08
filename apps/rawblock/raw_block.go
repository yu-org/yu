package rawblock

import (
	"bytes"
	"encoding/gob"
	"github.com/yu-org/yu/blockchain"
	"github.com/yu-org/yu/txn"
)

type RawBlock struct {
	BlockByt []byte
	TxnsByt  []byte
}

func NewRawBlock(block blockchain.IBlock, txns txn.SignedTxns) (*RawBlock, error) {
	blockByt, err := block.Encode()
	if err != nil {
		return nil, err
	}
	txnsByt, err := txns.Encode()
	if err != nil {
		return nil, err
	}
	return &RawBlock{
		BlockByt: blockByt,
		TxnsByt:  txnsByt,
	}, nil
}

func (r *RawBlock) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeRawBlock(data []byte) (*RawBlock, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var val RawBlock
	err := decoder.Decode(&val)
	return &val, err
}
