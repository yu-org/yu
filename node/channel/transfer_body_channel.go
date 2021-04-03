package channel

//
//import (
//	. "yu/node"
//	. "yu/storage/queue"
//	. "yu/utils/error_handle"
//)
//
//type MemTbChan struct {
//	c chan *TransferBody
//}
//
//func NewMemTbChan(cap int) TransferBodyChannel {
//	return &MemTbChan{c: make(chan *TransferBody, cap)}
//}
//
//func (mtc *MemTbChan) SendChan() chan<- *TransferBody {
//	return mtc.c
//}
//
//func (mtc *MemTbChan) RecvChan() <-chan *TransferBody {
//	return mtc.c
//}
//
//type DiskTbChan struct {
//	snd  chan *TransferBody
//	recv chan *TransferBody
//	q    Queue
//}
//
//func NewDiskTbChan(cap int, q Queue) TransferBodyChannel {
//	snd := make(chan *TransferBody, cap)
//	recv := make(chan *TransferBody, cap)
//	go LogfIfErr(q.PushAsync(TransferBodyTopic, snd), "push TransferBody into queue error: ")
//	go LogfIfErr(q.PopAsync(TransferBodyTopic, recv), "pop TransferBody from queue error: ")
//
//	return &DiskTbChan{
//		snd:  snd,
//		recv: recv,
//		q:    q,
//	}
//}
//
//func (dtc *DiskTbChan) SendChan() chan<- *TransferBody {
//	return dtc.snd
//}
//
//func (dtc *DiskTbChan) RecvChan() <-chan *TransferBody {
//	return dtc.recv
//}
