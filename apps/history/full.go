package history

import (
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
)

type FullHistory struct {
}

func (f *FullHistory) GetTripodHeader() *TripodHeader {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) CheckTxn(txn *SignedTxn) error {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) VerifyBlock(block *Block) bool {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) InitChain() {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) SyncHistory() {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) StartBlock(block *Block) {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) EndBlock(block *Block) {
	//TODO implement me
	panic("implement me")
}

func (f *FullHistory) FinalizeBlock(block *Block) {
	//TODO implement me
	panic("implement me")
}
