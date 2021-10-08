package poa

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/rawblock"
	. "github.com/yu-org/yu/blockchain"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/keypair"
	"github.com/yu-org/yu/node"
	. "github.com/yu-org/yu/tripod"
	"github.com/yu-org/yu/txn"
	"github.com/yu-org/yu/yerror"
	"time"
)

type Poa struct {
	meta     *TripodMeta
	authPool []PubKey

	localPubkey PubKey
	privKey     PrivKey
	packLimit   uint64

	msgChan chan []byte
}

func NewPoa(packLimit uint64, localPubkey PubKey, privKey PrivKey, authPool []PubKey) *Poa {

	return &Poa{
		meta:        NewTripodMeta("poa"),
		authPool:    authPool,
		packLimit:   packLimit,
		localPubkey: localPubkey,
		privKey:     privKey,
		msgChan:     make(chan []byte, 10),
	}
}

func (p *Poa) GetTripodMeta() *TripodMeta {
	return p.meta
}

func (p *Poa) Name() string {
	return p.meta.Name()
}

func (p *Poa) CheckTxn(txn *txn.SignedTxn) error {
	return nil
}

func (p *Poa) VerifyBlock(block IBlock, env *ChainEnv) bool {
	proposerPubkey := p.turnProposer(block)
	sig := block.GetSignature()
	return proposerPubkey.VerifySignature(block.GetHash().Bytes(), sig)
}

func (p *Poa) InitChain(env *ChainEnv, land *Land) error {
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

func (p *Poa) StartBlock(block IBlock, env *ChainEnv, land *Land) error {
	time.Sleep(2 * time.Second)

	proposer := p.turnProposer(block)
	if !proposer.Equals(p.localPubkey) {
		return nil
	}

	pool := env.Pool

	txns, err := pool.Pack(p.packLimit)
	if err != nil {
		return err
	}

	block.SetTxnsHashes(txn.FromArray(txns...).Hashes())

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return err
	}
	block.SetTxnRoot(txnRoot)
	block.SetHash(txnRoot)

	sig, err := p.privKey.SignData(txnRoot.Bytes())
	if err != nil {
		return err
	}
	block.SetSignature(sig)

	pool.Reset()
	err = env.Base.SetTxns(block.GetHash(), txns)
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

	return env.PubP2P(StartBlockTopic, rawBlockByt)
}

func (p *Poa) EndBlock(block IBlock, env *ChainEnv, land *Land) error {
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

func (p *Poa) FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error {
	return nil
}

// return TRUE if we use the p2p-block
func (p *Poa) UseBlocksFromP2P(block IBlock, env *ChainEnv, land *Land) bool {
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

func (p *Poa) useP2pBlock(msg []byte, block IBlock, env *ChainEnv, land *Land) bool {

	p2pRawBlock, err := rawblock.DecodeRawBlock(msg)
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
		logrus.Infof("Accept [LOCAL-P2P] block(%s) height(%d)", p2pBlock.GetHash().String(), p2pBlock.GetHeight())
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

func (p *Poa) turnProposer(block IBlock) PubKey {
	idx := block.GetHeight() % BlockNum(len(p.authPool))
	return p.authPool[idx]
}
