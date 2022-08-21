package poa

import (
	"bytes"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"go.uber.org/atomic"
	"time"
)

const BlockTime = 3

type Poa struct {
	*Tripod

	// key: crypto address, generate from pubkey
	validatorsMap map[Address]peer.ID
	myPubkey      PubKey
	myPrivKey     PrivKey

	validatorsList []Address

	currentHeight *atomic.Uint32

	recvChan chan *Block
	// local node index in addrs
	nodeIdx int
}

type ValidatorInfo struct {
	Pubkey PubKey
	P2pIP  string
}

func NewPoa(myPubkey PubKey, myPrivkey PrivKey, addrIps []ValidatorInfo) *Poa {
	header := NewTripod("Poa")

	var nodeIdx int

	validatorsAddr := make([]Address, 0)
	validators := make(map[Address]peer.ID)
	for _, addrIp := range addrIps {
		addr := addrIp.Pubkey.Address()
		p2pIP := addrIp.P2pIP

		peerID, err := peer.Decode(p2pIP)
		if err != nil {
			logrus.Fatalf("decode validatorIP(%s) error: %v", p2pIP, err)
		}
		validators[addr] = peerID

		if addr == myPubkey.Address() {
			nodeIdx = len(validatorsAddr)
		}

		validatorsAddr = append(validatorsAddr, addr)
	}

	h := &Poa{
		Tripod:         header,
		validatorsMap:  validators,
		validatorsList: validatorsAddr,
		myPubkey:       myPubkey,
		myPrivKey:      myPrivkey,
		currentHeight:  atomic.NewUint32(0),
		recvChan:       make(chan *Block, 10),
		nodeIdx:        nodeIdx,
	}
	h.SetInit(h)
	h.SetBlockCycle(h)
	h.SetBlockVerifier(h)
	h.SetTxnChecker(h)
	return h
}

func (h *Poa) ValidatorsP2pID() (peers []peer.ID) {
	for _, id := range h.validatorsMap {
		peers = append(peers, id)
	}
	return
}

func (h *Poa) LocalAddress() Address {
	return h.myPubkey.Address()
}

func (h *Poa) CheckTxn(txn *SignedTxn) error {
	return CheckSignature(txn)
}

func (h *Poa) VerifyBlock(block *Block) bool {
	minerPubkey, err := PubKeyFromBytes(block.MinerPubkey)
	if err != nil {
		logrus.Warnf("parse pubkey(%s) error: %v", block.MinerPubkey, err)
		return false
	}
	if _, ok := h.validatorsMap[minerPubkey.Address()]; !ok {
		logrus.Warn("illegal miner: ", minerPubkey.StringWithType())
		return false
	}
	return minerPubkey.VerifySignature(block.Hash.Bytes(), block.MinerSignature)
}

func (h *Poa) InitChain() {

	go func() {
		for {
			msg, err := h.P2pNetwork.SubP2P(StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			p2pBlock, err := DecodeBlock(msg)
			if err != nil {
				logrus.Error("decode p2pBlock from p2p error: ", err)
				continue
			}
			if bytes.Equal(p2pBlock.MinerPubkey, h.myPubkey.BytesWithType()) {
				continue
			}

			logrus.Debugf("accept block(%s), height(%d), miner(%s)",
				p2pBlock.Hash.String(), p2pBlock.Height, ToHex(p2pBlock.MinerPubkey))

			if h.getCurrentHeight() > p2pBlock.Height {
				continue
			}

			ok := h.VerifyBlock(p2pBlock)
			if !ok {
				logrus.Warnf("p2pBlock(%s) verify failed", p2pBlock.Hash.String())
				continue
			}

			h.recvChan <- p2pBlock
		}
	}()
}

func (h *Poa) StartBlock(block *Block) {
	now := time.Now()
	defer func() {
		duration := time.Since(now)
		time.Sleep(BlockTime*time.Second - duration)
	}()

	h.setCurrentHeight(block.Height)

	logrus.Info("====== start a new block ", block.Height)

	if h.AmILeader(block.Height) {
		if h.useP2pOrSkip(block) {
			logrus.Infof("--------USE P2P Height(%d) block(%s) miner(%s)",
				block.Height, block.Hash.String(), ToHex(block.MinerPubkey))
			return
		}
	}

	logrus.Info(" I am Leader! I mine the block! ")
	txns, err := h.Pool.Pack(3000)
	if err != nil {
		logrus.Panic("pack txns from pool: ", err)
	}
	logrus.Debugf("pack txns from pool: %d", len(txns))

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		logrus.Panic("make txn-root failed: ", err)
	}
	block.TxnRoot = txnRoot

	byt, _ := block.Encode()
	block.Hash = BytesToHash(Sha256(byt))

	// miner signs block
	block.MinerSignature, err = h.myPrivKey.SignData(block.Hash.Bytes())
	if err != nil {
		logrus.Panic("sign block failed: ", err)
	}
	block.MinerPubkey = h.myPubkey.BytesWithType()

	rawBlock := &Block{
		Header: block.Header,
		Txns:   txns,
	}

	h.State.StartBlock(block.Hash)

	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		logrus.Panic("encode raw-block failed: ", err)
	}

	err = h.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
	if err != nil {
		logrus.Panic("publish block to p2p failed: ", err)
	}
}

func (h *Poa) EndBlock(block *Block) {
	chain := h.Chain

	err := h.Execute(block)
	if err != nil {
		logrus.Panic("execute block failed: ", err)
	}

	err = chain.AppendBlock(block)
	if err != nil {
		logrus.Panic("append block failed: ", err)
	}

	err = h.Pool.Reset(block.Txns)
	if err != nil {
		logrus.Panic("reset pool failed: ", err)
	}

	logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
		Info("append block")

	h.State.FinalizeBlock(block.Hash)
}

func (h *Poa) FinalizeBlock(block *Block) {
	logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
		Info("finalize block")
	h.Chain.Finalize(block.Hash)
}

func (h *Poa) CompeteLeader(blockHeight BlockNum) Address {
	idx := (int(blockHeight) - 1) % len(h.validatorsList)
	leader := h.validatorsList[idx]
	logrus.Debugf("compete a leader(%s) in round(%d)", leader.String(), blockHeight)
	return leader
}

func (h *Poa) AmILeader(blockHeight BlockNum) bool {
	return h.CompeteLeader(blockHeight) == h.LocalAddress()
}

func (h *Poa) useP2pOrSkip(localBlock *Block) bool {
LOOP:
	select {
	case p2pBlock := <-h.recvChan:
		if h.getCurrentHeight() > p2pBlock.Height {
			goto LOOP
		}
		localBlock.CopyFrom(p2pBlock)
		h.State.StartBlock(localBlock.Hash)
		return true
	case <-time.NewTicker(h.calulateWaitTime(localBlock)).C:
		return false
	}
}

func (h *Poa) calulateWaitTime(block *Block) time.Duration {
	height := int(block.Height)
	shouldLeaderIdx := (height - 1) % len(h.validatorsList)
	n := shouldLeaderIdx - h.nodeIdx
	if n < 0 {
		n = -n
	}

	return time.Duration(BlockTime+n) * time.Second
}

func (h *Poa) getCurrentHeight() BlockNum {
	return BlockNum(h.currentHeight.Load())
}

func (h *Poa) setCurrentHeight(height BlockNum) {
	h.currentHeight.Store(uint32(height))
}
