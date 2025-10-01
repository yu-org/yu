package types

import (
	"github.com/yu-org/yu/apps/eth/ethrpc"
)

type CallRequest struct {
	TxArgs *ethrpc.TransactionArgs `json:"tx_args"`
}

type CallResponse struct {
	Ret []byte `json:"ret"`
	Err error  `json:"err"`
}
