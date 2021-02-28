package pow

import (
	"math/big"
	. "yu/blockchain"
	spow "yu/consensus/pow"
	"yu/context"
	. "yu/tripod"
	"yu/txn"
)

type Pow struct {
	meta       *TripodMeta
	target     *big.Int
	targetBits int64
}

func NewPow() *Pow {
	meta := NewTripodMeta("pow")
	var targetBits int64 = 16
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &Pow{
		meta:       meta,
		targetBits: targetBits,
	}
}

func (p *Pow) TripodMeta() *TripodMeta {
	return p.meta
}

func (*Pow) CheckTxn(txn.IsignedTxn) error {
	return nil
}

func (*Pow) StartBlock(*context.Context, IBlock) error {
	return nil
}

func (p *Pow) ExecuteTxns(_ *context.Context, block IBlock, txns []txn.IsignedTxn) error {
	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return err
	}
	block.SetHash(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return err
	}
	block.SetExtra(nonce)
	block.SetHash(hash)
	return nil
}

func (*Pow) EndBlock(*context.Context, IBlock) error {
	return nil
}

func (*Pow) FinalizeBlock(*context.Context, IBlock) error {
	return nil
}
