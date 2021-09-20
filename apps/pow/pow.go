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
	"github.com/yu-altar/yu/yerror"
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

func (p *Pow) StartBlock(block IBlock, env *ChainEnv, land *Land, msgChan <-chan []byte) ([]byte, error) {
	time.Sleep(2 * time.Second)

	chain := env.Chain
	pool := env.Pool

	logrus.Info("start block...................")

	prevBlock, err := chain.GetEndBlock()
	if err != nil {
		return nil, err
	}

	logrus.Infof("prev-block hash is (%s), height is (%d)", block.GetPrevHash().String(), block.GetHeight()-1)

	block.(*Block).SetChainLen(prevBlock.(*Block).ChainLen + 1)

	if p.UseBlocksFromP2P(block, msgChan, env, land) {
		logrus.Infof("USE P2P block(%s)", block.GetHash().String())
		return nil, nil
	}

	txns, err := pool.Pack(p.pkgTxnsLimit)
	if err != nil {
		return nil, err
	}

	hashes := txn.FromArray(txns...).Hashes()
	block.SetTxnsHashes(hashes)

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return nil, err
	}
	block.SetTxnRoot(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return nil, err
	}

	env.Pool.Reset()

	block.(*Block).SetNonce(uint64(nonce))
	block.SetHash(hash)

	block.SetPeerID(env.P2pID)

	env.StartBlock(hash)
	err = env.Base.SetTxns(block.GetHash(), txns)
	if err != nil {
		return nil, err
	}

	return block.Encode()
}

func (*Pow) EndBlock(block IBlock, env *ChainEnv, land *Land, _ <-chan []byte) ([]byte, error) {
	chain := env.Chain
	pool := env.Pool

	err := node.ExecuteTxns(block, env, land)
	if err != nil {
		return nil, err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return nil, err
	}

	logrus.Infof("append block(%d)", block.GetHeight())

	env.SetCanRead(block.GetHash())

	return nil, pool.Flush()
}

func (*Pow) FinalizeBlock(_ IBlock, _ *ChainEnv, _ *Land, _ <-chan []byte) ([]byte, error) {
	return nil, nil
}

// return TRUE if we use the p2p-block
func (*Pow) UseBlocksFromP2P(block IBlock, msgChan <-chan []byte, env *ChainEnv, land *Land) (useIt bool) {
	for i := 0; i < len(msgChan); i++ {

		msg := <-msgChan
		p2pBlock, err := env.Chain.NewEmptyBlock().Decode(msg)
		if err != nil {
			logrus.Error("decode p2p-block error: ", err)
			return false
		}
		err = land.RangeList(func(tri Tripod) error {
			if tri.VerifyBlock(p2pBlock, env) {
				return nil
			}
			return yerror.BlockIllegal(p2pBlock.GetHash())
		})

		if err != nil {
			logrus.Error("verify p2p-block error: ", err)
			return false
		}

		if p2pBlock.GetHeight() < block.GetHeight() {
			err = env.Chain.AppendBlock(p2pBlock)
			if err != nil {
				logrus.Errorf("append p2p-block(%s) error: %s", p2pBlock.GetHash().String(), err.Error())
			}
		}

		if p2pBlock.GetHeight() == block.GetHeight() {
			block.CopyFrom(p2pBlock)
			env.StartBlock(block.GetHash())
			return true
		}
	}

	return false
}
