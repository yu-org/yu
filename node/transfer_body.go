package node

import (
	"encoding/json"
	. "yu/blockchain"
	. "yu/common"
	. "yu/txn"
)

type BodyType int

const (
	BlockTransfer BodyType = iota
	TxnsTransfer
)

type TransferBody struct {
	Type BodyType   `json:"type"`
	Body JsonString `json:"body"`
}

func NewBlockTransferBody(block IBlock) (*TransferBody, error) {
	byt, err := block.Encode()
	if err != nil {
		return nil, err
	}
	return &TransferBody{
		Type: BlockTransfer,
		Body: JsonString(byt),
	}, nil
}

func NewTxnsTransferBody(txns SignedTxns) (*TransferBody, error) {
	byt, err := txns.Encode()
	if err != nil {
		return nil, err
	}
	return &TransferBody{
		Type: TxnsTransfer,
		Body: JsonString(byt),
	}, nil
}

func (tb *TransferBody) Encode() ([]byte, error) {
	return json.Marshal(tb)
}

func DecodeTb(data []byte) (*TransferBody, error) {
	var tb TransferBody
	err := json.Unmarshal(data, &tb)
	return &tb, err
}
