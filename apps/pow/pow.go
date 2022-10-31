package pow

//
//import (
//	"github.com/sirupsen/logrus"
//	. "github.com/yu-org/yu/common"
//	spow "github.com/yu-org/yu/consensus/pow"
//	. "github.com/yu-org/yu/core/chain_env"
//	. "github.com/yu-org/yu/core/keypair"
//	. "github.com/yu-org/yu/core/tripod"
//	"github.com/yu-org/yu/core/types"
//	"math/big"
//	"time"
//)
//
//type Pow struct {
//	meta       *Tripod
//	target     *big.Int
//	targetBits int64
//
//	myPrivKey PrivKey
//	myPubkey  PubKey
//
//	env *ChainEnv
//
//	packLimit uint64
//	blockTick *time.Ticker
//	p2pTick   *time.Ticker
//	msgChan   chan []byte
//}
//
//func NewPow(packLimit uint64) *Pow {
//	meta := NewTripod("pow")
//	var targetBits int64 = 16
//	target := big.NewInt(1)
//	target.Lsh(target, uint(256-targetBits))
//
//	pubkey, privkey, err := GenKeyPair(Sr25519)
//	if err != nil {
//		logrus.Fatalf("generate my keypair error: %s", err.Error())
//	}
//
//	return &Pow{
//		meta:       meta,
//		target:     target,
//		targetBits: targetBits,
//		myPrivKey:  privkey,
//		myPubkey:   pubkey,
//
//		packLimit: packLimit,
//		blockTick: time.NewTicker(time.Second * 2),
//		p2pTick:   time.NewTicker(time.Second),
//		msgChan:   make(chan []byte, 100),
//	}
//}
//
//func (p *Pow) GetTripodHeader() *Tripod {
//	return p.meta
//}
//
//func (p *Pow) Name() string {
//	return p.meta.Name()
//}
//
//func (p *Pow) SetChainEnv(env *ChainEnv) {
//	p.env = env
//}
//
//func (*Pow) CheckTxn(*types.SignedTxn) error {
//	return nil
//}
//
//func (p *Pow) VerifyBlock(block *types.CompactBlock) bool {
//	return spow.Validate(block, p.target, p.targetBits)
//}
//
//func (p *Pow) InitChain() error {
//	chain := p.env.Chain
//	gensisBlock := &types.CompactBlock{
//		Header: &types.Header{},
//	}
//	err := chain.SetGenesis(gensisBlock)
//	if err != nil {
//		return err
//	}
//	go func() {
//		for {
//			msg, err := p.env.P2pNetwork.SubP2P(StartBlockTopic)
//			if err != nil {
//				logrus.Error("subscribe message from P2P error: ", err)
//				continue
//			}
//			p.msgChan <- msg
//		}
//
//	}()
//	return nil
//}
//
//func (p *Pow) StartBlock(block *types.CompactBlock) error {
//	time.Sleep(2 * time.Second)
//
//	pool := p.env.Pool
//
//	logrus.Info("start block...................")
//
//	logrus.Infof("prev-block hash is (%s), height is (%d)", block.PrevHash.String(), block.Height-1)
//
//	if p.UseBlocksFromP2P(block) {
//		logrus.Infof("--------USE P2P block(%s)", block.Hash.String())
//		return nil
//	}
//
//	txns, err := pool.Pack(p.packLimit)
//	if err != nil {
//		return err
//	}
//
//	hashes := types.FromArray(txns...).Hashes()
//	block.TxnsHashes = hashes
//
//	txnRoot, err := types.MakeTxnRoot(txns)
//	if err != nil {
//		return err
//	}
//	block.TxnRoot = txnRoot
//
//	nonce, hash, err := spow.Run(block, p.target, p.targetBits)
//	if err != nil {
//		return err
//	}
//
//	block.Nonce = uint64(nonce)
//	block.Hash = hash
//
//	p.env.State.StartBlock(hash)
//
//	rawBlock := &types.Block{
//		CompactBlock: block,
//		Txns:         txns,
//	}
//
//	rawBlockByt, err := rawBlock.Encode()
//	if err != nil {
//		return err
//	}
//
//	return p.env.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
//}
//
//func (p *Pow) EndBlock(block *types.CompactBlock) error {
//	chain := p.env.Chain
//	pool := p.env.Pool
//
//	err := p.env.Execute(block)
//	if err != nil {
//		return err
//	}
//
//	err = chain.AppendCompactBlock(block)
//	if err != nil {
//		return err
//	}
//
//	err = pool.Reset(block)
//	if err != nil {
//		return err
//	}
//
//	logrus.Infof("append block(%d) (%s)", block.Height, block.Hash.String())
//
//	p.env.State.FinalizeBlock(block.Hash)
//
//	return nil
//}
//
//func (*Pow) FinalizeBlock(_ *types.CompactBlock) error {
//	return nil
//}
//
//// return TRUE if we use the p2p-block
//func (p *Pow) UseBlocksFromP2P(block *types.CompactBlock) bool {
//	msgCount := len(p.msgChan)
//	if msgCount > 0 {
//		for i := 0; i < msgCount; i++ {
//			msg := <-p.msgChan
//			if p.useP2pBlock(msg, block) {
//				return true
//			}
//		}
//	}
//	return false
//}
//
//func (p *Pow) useP2pBlock(msg []byte, block *types.CompactBlock) bool {
//
//	p2pRawBlock, err := types.DecodeBlock(msg)
//	if err != nil {
//		logrus.Error("decode p2p-raw-block error: ", err)
//		return false
//	}
//
//	p2pBlock := p2pRawBlock.CompactBlock
//
//	if p2pBlock.PeerID == block.PeerID {
//		logrus.Infof("Accept [LOCAL-P2P] block(%s) height(%d)", p2pBlock.Hash.String(), p2pBlock.Height)
//		return false
//	}
//
//	logrus.Infof("Accept [P2P] block(%s) height(%d)", p2pBlock.Hash.String(), p2pBlock.Height)
//
//	if p2pBlock.Height == block.Height {
//		if !p.VerifyBlock(p2pBlock) {
//			logrus.Error("verify p2p-block error: ", err)
//			return false
//		}
//
//		block.CopyFrom(p2pBlock)
//		stxns := p2pRawBlock.Txns
//		err = p.env.TxDB.SetTxns(block.Hash, stxns)
//		if err != nil {
//			logrus.Error("set txns of p2p-block into base error: ", err)
//			return false
//		}
//		p.env.State.StartBlock(block.Hash)
//		err = p.env.Pool.Reset(block)
//		if err != nil {
//			logrus.Error("clear txpool error: ", err)
//			return false
//		}
//		return true
//	}
//
//	return false
//}
