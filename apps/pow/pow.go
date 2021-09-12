package pow

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-altar/yu/blockchain"
	. "github.com/yu-altar/yu/chain_env"
	spow "github.com/yu-altar/yu/consensus/pow"
	. "github.com/yu-altar/yu/keypair"
	"github.com/yu-altar/yu/node"
	. "github.com/yu-altar/yu/tripod"
	"github.com/yu-altar/yu/txn"
	"math/big"
	"time"
)

type Pow struct {
	meta       *TripodMeta
	target     *big.Int
	targetBits int64

	myPrivKey PrivKey
	myPubkey  PubKey

	pkgTxnsLimit uint64
}

func NewPow(pkgTxnsLimit uint64) *Pow {
	meta := NewTripodMeta("pow")
	var targetBits int64 = 16
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pubkey, privkey, err := GenKeyPair(Sr25519)
	if err != nil {
		logrus.Fatalf("generate my keypair error: %s", err.Error())
	}

	return &Pow{
		meta:       meta,
		target:     target,
		targetBits: targetBits,
		myPrivKey:  privkey,
		myPubkey:   pubkey,

		pkgTxnsLimit: pkgTxnsLimit,
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

func (p *Pow) InitChain(env *ChainEnv, _ *Land) error {
	chain := env.Chain
	gensisBlock := &Block{
		Header: &Header{},
	}
	return chain.SetGenesis(gensisBlock)
}

func (p *Pow) StartBlock(block IBlock, env *ChainEnv, _ *Land) (needBroadcast bool, err error) {
	time.Sleep(2 * time.Second)

	chain := env.Chain
	pool := env.Pool

	logrus.Info("start block...................")

	prevBlock, err := chain.GetEndBlock()
	if err != nil {
		return
	}

	logrus.Infof("prev-block hash is (%s), height is (%d)", block.GetPrevHash().String(), block.GetHeight()-1)

	block.(*Block).SetChainLen(prevBlock.(*Block).ChainLen + 1)

	pbMap, err := chain.TakeP2pBlocksBefore(block.GetHeight())
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

	pbsht, err := chain.TakeP2pBlocks(block.GetHeight())
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

	block.SetPeerID(env.P2pID)

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
