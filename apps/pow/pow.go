package pow

import (
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
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

func (*Pow) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (p *Pow) ValidateBlock(b IBlock) bool {
	return spow.Validate(b, p.target, p.targetBits)
}

func (*Pow) InitChain(chain IBlockChain, _ IBlockBase) error {
	gensisBlock := chain.NewDefaultBlock()
	gensisBlock.SetHeight(0)
	return chain.SetGenesis(gensisBlock)
}

func (p *Pow) StartBlock(chain IBlockChain, block IBlock, pool txpool.ItxPool) (needBroadcast bool, err error) {
	time.Sleep(2 * time.Second)
	chains, err := chain.Longest()
	if err != nil {
		return
	}

	for _, ch := range chains {
		_ = ch.Range(func(block IBlock) error {
			logrus.Info("longest chain block is ", block.GetHeader().GetHash().String())
			return nil
		})
	}

	preBlock := chains[0].Last()

	prevHeight := preBlock.GetHeader().GetHeight()
	prevHash := preBlock.GetHeader().GetHash()

	height := prevHeight + 1

	p2pBlocks, err := chain.GetBlocksFromP2P(height)
	if err != nil {
		logrus.Errorf("get p2p-blocks error: %s", err.Error())
	}
	if len(p2pBlocks) > 0 {
		block = p2pBlocks[0]
		return
	}

	needBroadcast = true

	block.SetPreHash(prevHash)
	block.SetHeight(height)

	txns, err := pool.Package("", p.pkgTxnsLimit)
	if err != nil {
		return
	}
	txnsHashes := make([]Hash, 0)
	for _, hash := range txnsHashes {
		txnsHashes = append(txnsHashes, hash)
	}
	block.SetTxnsHashes(txnsHashes)

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return
	}
	block.SetStateRoot(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return
	}
	block.(*Block).SetNonce(nonce)
	block.SetHash(hash)

	return
}

func (*Pow) EndBlock(chain IBlockChain, block IBlock) error {
	return chain.AppendBlock(block)
}

func (*Pow) FinalizeBlock(IBlockChain, IBlock) error {
	return nil
}
