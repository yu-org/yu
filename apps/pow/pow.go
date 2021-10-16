package pow

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/rawblock"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	spow "github.com/yu-org/yu/consensus/pow"
	. "github.com/yu-org/yu/keypair"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/txn"
	"math/big"
	"time"
)

type Pow struct {
	meta       *TripodMeta
	target     *big.Int
	targetBits int64

	myPrivKey PrivKey
	myPubkey  PubKey

	env *ChainEnv

	packLimit uint64
	blockTick *time.Ticker
	p2pTick   *time.Ticker
	msgChan   chan []byte
}

func NewPow(packLimit uint64) *Pow {
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

		packLimit: packLimit,
		blockTick: time.NewTicker(time.Second * 2),
		p2pTick:   time.NewTicker(time.Second),
		msgChan:   make(chan []byte, 100),
	}
}

func (p *Pow) GetTripodMeta() *TripodMeta {
	return p.meta
}

func (p *Pow) Name() string {
	return p.meta.Name()
}

func (p *Pow) SetChainEnv(env *ChainEnv) {
	p.env = env
}

func (*Pow) CheckTxn(*txn.SignedTxn) error {
	return nil
}

func (p *Pow) VerifyBlock(block IBlock) bool {
	return spow.Validate(block, p.target, p.targetBits)
}

func (p *Pow) InitChain() error {
	chain := p.env.Chain
	gensisBlock := &Block{
		Header: &Header{},
	}
	err := chain.SetGenesis(gensisBlock)
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := p.env.SubP2P(StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			p.msgChan <- msg
		}

	}()
	return nil
}

func (p *Pow) StartBlock(block IBlock) error {
	time.Sleep(2 * time.Second)

	pool := p.env.Pool

	logrus.Info("start block...................")

	logrus.Infof("prev-block hash is (%s), height is (%d)", block.GetPrevHash().String(), block.GetHeight()-1)

	if p.UseBlocksFromP2P(block) {
		logrus.Infof("--------USE P2P block(%s)", block.GetHash().String())
		return nil
	}

	txns, err := pool.Pack(p.packLimit)
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

	block.(*Block).SetNonce(uint64(nonce))
	block.SetHash(hash)

	p.env.StartBlock(hash)
	err = p.env.Base.SetTxns(block.GetHash(), txns)
	if err != nil {
		return err
	}

	rawBlock, err := rawblock.NewRawBlock(block, txns)
	if err != nil {
		return err
	}
	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		return err
	}

	return p.env.PubP2P(StartBlockTopic, rawBlockByt)
}

func (p *Pow) EndBlock(block IBlock) error {
	chain := p.env.Chain
	pool := p.env.Pool

	err := p.env.Execute(block)
	if err != nil {
		return err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	logrus.Infof("append block(%d) (%s)", block.GetHeight(), block.GetHash().String())

	p.env.SetCanRead(block.GetHash())

	return pool.Reset()
}

func (*Pow) FinalizeBlock(_ IBlock) error {
	return nil
}

// return TRUE if we use the p2p-block
func (p *Pow) UseBlocksFromP2P(block IBlock) bool {
	msgCount := len(p.msgChan)
	if msgCount > 0 {
		for i := 0; i < msgCount; i++ {
			msg := <-p.msgChan
			if p.useP2pBlock(msg, block) {
				return true
			}
		}
	}
	return false
}

func (p *Pow) useP2pBlock(msg []byte, block IBlock) bool {

	p2pRawBlock, err := rawblock.DecodeRawBlock(msg)
	if err != nil {
		logrus.Error("decode p2p-raw-block error: ", err)
		return false
	}

	p2pBlock, err := p.env.Chain.NewEmptyBlock().Decode(p2pRawBlock.BlockByt)
	if err != nil {
		logrus.Error("decode p2p-block error: ", err)
		return false
	}

	if p2pBlock.GetPeerID() == block.GetPeerID() {
		logrus.Infof("Accept [LOCAL-P2P] block(%s) height(%d)", p2pBlock.GetHash().String(), p2pBlock.GetHeight())
		return false
	}

	logrus.Infof("Accept [P2P] block(%s) height(%d)", p2pBlock.GetHash().String(), p2pBlock.GetHeight())

	if p2pBlock.GetHeight() == block.GetHeight() {
		if !p.VerifyBlock(p2pBlock) {
			logrus.Error("verify p2p-block error: ", err)
			return false
		}

		block.CopyFrom(p2pBlock)
		stxns, err := txn.DecodeSignedTxns(p2pRawBlock.TxnsByt)
		if err != nil {
			logrus.Error("decode txns of p2p-block error: ", err)
			return false
		}
		err = p.env.Base.SetTxns(block.GetHash(), stxns)
		if err != nil {
			logrus.Error("set txns of p2p-block into base error: ", err)
			return false
		}
		p.env.StartBlock(block.GetHash())
		err = p.env.Pool.RemoveTxns(block.GetTxnsHashes())
		if err != nil {
			logrus.Error("clear txpool error: ", err)
			return false
		}
		return true
	}

	return false
}
