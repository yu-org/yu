package pow

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	spow "github.com/yu-org/yu/consensus/pow"
	. "github.com/yu-org/yu/keypair"
	"github.com/yu-org/yu/node"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/txn"
	"github.com/yu-org/yu/yerror"
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
	blockTick    *time.Ticker
	p2pTick      *time.Ticker
	msgChan      chan []byte
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
		blockTick:    time.NewTicker(time.Second * 2),
		p2pTick:      time.NewTicker(time.Second),
		msgChan:      make(chan []byte, 100),
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
	err := chain.SetGenesis(gensisBlock)
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := env.SubP2P(StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			p.msgChan <- msg
		}

	}()
	return nil
}

func (p *Pow) StartBlock(block IBlock, env *ChainEnv, land *Land) error {
	time.Sleep(2 * time.Second)

	chain := env.Chain
	pool := env.Pool

	logrus.Info("start block...................")

	prevBlock, err := chain.GetEndBlock()
	if err != nil {
		return err
	}

	logrus.Infof("prev-block hash is (%s), height is (%d)", block.GetPrevHash().String(), block.GetHeight()-1)

	block.(*Block).SetChainLen(prevBlock.(*Block).ChainLen + 1)

	if p.UseBlocksFromP2P(block, env, land) {
		logrus.Infof("--------USE P2P block(%s)", block.GetHash().String())
		return nil
	}

	txns, err := pool.Pack(p.pkgTxnsLimit)
	if err != nil {
		return err
	}

	hashes := txn.FromArray(txns...).Hashes()
	block.SetTxnsHashes(hashes)

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return err
	}
	block.SetTxnRoot(txnRoot)

	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
	if err != nil {
		return err
	}

	env.Pool.Reset()

	block.(*Block).SetNonce(uint64(nonce))
	block.SetHash(hash)

	env.StartBlock(hash)
	err = env.Base.SetTxns(block.GetHash(), txns)
	if err != nil {
		return err
	}

	rawBlock, err := NewRawBlock(block, txns)
	if err != nil {
		return err
	}
	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		return err
	}

	return env.PubP2P(StartBlockTopic, rawBlockByt)
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

	logrus.Infof("append block(%d) (%s)", block.GetHeight(), block.GetHash().String())

	env.SetCanRead(block.GetHash())

	return pool.Flush()
}

func (*Pow) FinalizeBlock(_ IBlock, _ *ChainEnv, _ *Land) error {
	return nil
}

// return TRUE if we use the p2p-block
func (p *Pow) UseBlocksFromP2P(block IBlock, env *ChainEnv, land *Land) bool {
	msgCount := len(p.msgChan)
	if msgCount > 0 {
		for i := 0; i < msgCount; i++ {
			msg := <-p.msgChan
			if p.useP2pBlock(msg, block, env, land) {
				return true
			}
		}
	}
	return false
}

func (p *Pow) useP2pBlock(msg []byte, block IBlock, env *ChainEnv, land *Land) bool {

	p2pRawBlock, err := DecodeRawBlock(msg)
	if err != nil {
		logrus.Error("decode p2p-raw-block error: ", err)
		return false
	}

	p2pBlock, err := env.Chain.NewEmptyBlock().Decode(p2pRawBlock.BlockByt)
	if err != nil {
		logrus.Error("decode p2p-block error: ", err)
		return false
	}

	if p2pBlock.GetPeerID() == block.GetPeerID() {
		return false
	}

	logrus.Infof("Accept [P2P] block(%s) height(%d)", p2pBlock.GetHash().String(), p2pBlock.GetHeight())

	if p2pBlock.GetHeight() == block.GetHeight() {
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

		block.CopyFrom(p2pBlock)
		stxns, err := txn.DecodeSignedTxns(p2pRawBlock.TxnsByt)
		if err != nil {
			logrus.Error("decode txns of p2p-block error: ", err)
			return false
		}
		err = env.Base.SetTxns(block.GetHash(), stxns)
		if err != nil {
			logrus.Error("set txns of p2p-block into base error: ", err)
			return false
		}
		env.StartBlock(block.GetHash())
		err = env.Pool.RemoveTxns(block.GetTxnsHashes())
		if err != nil {
			logrus.Error("clear txpool error: ", err)
			return false
		}
		return true
	}

	return false
}
