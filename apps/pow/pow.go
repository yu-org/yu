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

	pkgTxnsLimit uint64
}

func NewPow(pkgTxnsLimit uint64) *Pow {
	meta := NewTripodMeta("pow")
	var targetBits int64 = 16
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &Pow{
		meta:         meta,
		target:       target,
		targetBits:   targetBits,
		pkgTxnsLimit: pkgTxnsLimit,
	}
}

func (p *Pow) TripodMeta() *TripodMeta {
	return p.meta
}

func (*Pow) CheckTxn(txn.IsignedTxn) error {
	return nil
}

func (p *Pow) StartBlock(chain IBlockChain, block IBlock, pool txpool.ItxPool) error {
	chains, err := chain.Longest()
	if err != nil {
		return err
	}

	preBlock := chains[0].Last()

	height := preBlock.Header().Height()
	preHash := preBlock.Header().PrevHash()

	block.SetPreHash(preHash)
	block.SetBlockNumber(height + 1)

	txns, err := pool.Package("", p.pkgTxnsLimit)
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
	block.SetStateRoot(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return err
	}
	block.SetExtra(nonce)
	block.SetHash(hash)

	return nil
}

func (*Pow) EndBlock(chain IBlockChain, block IBlock) error {
	return chain.AppendBlock(block)
}

func (*Pow) FinalizeBlock(IBlockChain, IBlock) error {
	return nil
}
