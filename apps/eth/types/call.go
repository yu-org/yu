package types

type CallRequest struct {
	TxArgs *TransactionArgs `json:"tx_args"`
}

type CallResponse struct {
	Ret []byte `json:"ret"`
	Err error  `json:"err"`
}
