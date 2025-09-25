package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type CallRequest struct {
	Input    []byte         `json:"input"`
	Address  common.Address `json:"address"`
	Origin   common.Address `json:"origin"`
	GasLimit uint64         `json:"gasLimit"`
	GasPrice *big.Int       `json:"gasPrice"`
	Value    *big.Int       `json:"value"`
}

type CallResponse struct {
	Ret         []byte `json:"ret"`
	LeftOverGas uint64 `json:"leftOverGas"`
	Err         error  `json:"err"`
}
