package channel

//import (
//	. "yu/blockchain"
//	. "yu/node"
//	. "yu/txn"
//)
//
//type BlockChannel interface {
//	SendChan() chan<- IBlock
//	RecvChan() <-chan IBlock
//}
//
//type TxnsChannel interface {
//	SendChan() chan<- IsignedTxn
//	RecvChan() <-chan IsignedTxn
//}
//
//type TransferBodyChannel interface {
//	SendChan() chan<- *TransferBody
//	RecvChan() <-chan *TransferBody
//}
//
//const (
//	BlockTopic        = "block"
//	TxnsTopic         = "txns"
//	TransferBodyTopic = "transfer-body"
//)
