package evm

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/yu-org/yu/apps/eth/types"
)

type TxRequest struct {
	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`

	TxArgs *types.TransactionArgs `json:"tx_args"`

	IsInternalCall bool `json:"is_internal_call"`
}

func (t *TxRequest) ToEthTx() *ethtypes.Transaction {
	return t.TxArgs.ToTransaction(t.V, t.R, t.S)
}

func (t *TxRequest) Encode() ([]byte, error) {
	// Use JSON encoding instead of gob to handle hexutil.Big properly
	return json.Marshal(t)
}

func DecodeTxReq(b []byte) (*TxRequest, error) {
	txReq := new(TxRequest)

	err := json.Unmarshal(b, txReq)
	if err != nil {
		return nil, err
	}
	return txReq, nil
}

type CreateRequest struct {
	Input    []byte         `json:"input"`
	Origin   common.Address `json:"origin"`
	GasLimit uint64         `json:"gasLimit"`
	GasPrice *big.Int       `json:"gasPrice"`
	Value    *big.Int       `json:"value"`
}
