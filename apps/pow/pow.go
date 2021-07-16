package pow

import (
	. "github.com/Lawliet-Chan/yu/blockchain"
	. "github.com/Lawliet-Chan/yu/chain_env"
	spow "github.com/Lawliet-Chan/yu/consensus/pow"
	"github.com/Lawliet-Chan/yu/node"
	. "github.com/Lawliet-Chan/yu/tripod"
	"github.com/Lawliet-Chan/yu/txn"
	ytime "github.com/Lawliet-Chan/yu/utils/time"
	"github.com/sirupsen/logrus"
	"math/big"
	"time"
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

func newDefaultBlock() *Block {
	return &Block{
		Header: &Header{
			Timestamp: ytime.NowNanoTsU64(),
		},
	}
}

func (p *Pow) GetTripodMeta() *TripodMeta {
	return p.meta
}

func (p *Pow) Name() string {
	return p.meta.Name()
}

func (*Pow) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (p *Pow) VerifyBlock(block IBlock, _ *ChainEnv) bool {
	return spow.Validate(block, p.target, p.targetBits)
}

func (*Pow) InitChain(env *ChainEnv, _ *Land) error {
	chain := env.Chain
	gensisBlock := &Block{
		Header: &Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (p *Pow) StartBlock(block IBlock, env *ChainEnv, _ *Land) (needBroadcast bool, err error) {
	time.Sleep(2 * time.Second)

	block.CopyFrom(newDefaultBlock())
	chain := env.Chain
	pool := env.Pool

	prevBlock, err := chain.GetEndBlock()
	if err != nil {
		return
	}

	logrus.Info("start block...................")

	prevHeight := prevBlock.GetHeight()
	prevHash := prevBlock.GetHash()

	logrus.Infof("prev-block hash is (%s), height is (%d)", prevHash.String(), prevHeight)

	height := prevHeight + 1

	pbMap, err := chain.TakeP2pBlocksBefore(height)
	if err != nil {
		logrus.Errorf("get p2p-blocks before error: %s", err.Error())
	}

	for _, pbs := range pbMap {
		for _, pb := range pbs {
			err = chain.AppendBlock(pb)
			if err != nil {
				return
			}
		}
	}

	pbsht, err := chain.TakeP2pBlocks(height)
	if err != nil {
		logrus.Errorf("get p2p-blocks error: %s", err.Error())
	}
	if len(pbsht) > 0 {
		block.CopyFrom(pbsht[0])
		logrus.Infof("USE P2P block(%s)", block.GetHash().String())
		env.StartBlock(block.GetHash())
		return
	}

	needBroadcast = true

	block.SetPreHash(prevHash)
	block.SetHeight(height)

	txns, err := pool.Pack(p.pkgTxnsLimit)
	if err != nil {
		return
	}

	hashes := txn.FromArray(txns...).Hashes()
	block.SetTxnsHashes(hashes)

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return
	}
	block.SetTxnRoot(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return
	}

	env.Pool.Reset()

	block.(*Block).SetNonce(uint64(nonce))
	block.SetHash(hash)

	block.SetProducerPeer(env.Peer.ID())

	env.StartBlock(hash)
	err = env.Base.SetTxns(block.GetHash(), txns)
	return
}

func (*Pow) EndBlock(block IBlock, env *ChainEnv, land *Land) error {
	chain := env.Chain
	pool := env.Pool

	err := node.ExecuteTxns(block, env, land)
	if err != nil {
		return err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	logrus.Infof("append block(%d)", block.GetHeight())

	env.SetCanRead(block.GetHash())

	return pool.Flush()
}

func (*Pow) FinalizeBlock(_ IBlock, _ *ChainEnv, _ *Land) error {
	return nil
}
