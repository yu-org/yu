package pow

import (
	"math/big"
	. "yu/blockchain"
	. "yu/common"
	spow "yu/consensus/pow"
	. "yu/tripod"
	"yu/txn"
	"yu/txpool"
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

func (*Pow) StartBlock(IBlock) error {
	return nil
}

func (p *Pow) HandleTxns(block IBlock, pool txpool.ItxPool) error {
	txns, err := pool.Package(1024)
	if err != nil {
		return err
	}
	txnsHashes := make([]Hash, 0)
	for _, hash := range txnsHashes {
		txnsHashes = append(txnsHashes, hash)
	}
	block.SetTxnsHashes(txnsHashes)

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

func (*Pow) EndBlock(IBlock) error {
	return nil
}

func (*Pow) FinalizeBlock(IBlock) error {
	return nil
}
