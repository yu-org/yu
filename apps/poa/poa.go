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
	header *TripodHeader

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
	header := NewTripodHeader("Poa")

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
		header:         header,
		validatorsMap:  validators,
		validatorsList: validatorsAddr,
		myPubkey:       myPubkey,
		myPrivKey:      myPrivkey,
		currentHeight:  atomic.NewUint32(0),
		recvChan:       make(chan *Block, 10),
		nodeIdx:        nodeIdx,
	}
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

func (h *Poa) GetTripodHeader() *TripodHeader {
	return h.header
}

func (h *Poa) CheckTxn(txn *SignedTxn) error {
	return CheckSignature(txn)
}

func (h *Poa) VerifyBlock(block *CompactBlock) bool {
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

func (h *Poa) InitChain() error {
	rootPubkey, rootPrivkey := GenSrKey([]byte("root"))
	genesisHash := HexToHash("genesis")
	signer, err := rootPrivkey.SignData(genesisHash.Bytes())
	if err != nil {
		return err
	}

	chain := h.GetTripodHeader().ChainEnv.Chain

	gensisBlock := &CompactBlock{
		Header: &Header{
			Hash:           genesisHash,
			MinerPubkey:    rootPubkey.BytesWithType(),
			MinerSignature: signer,
		},
	}

	err = chain.SetGenesis(gensisBlock)
	if err != nil {
		return err
	}
	err = chain.Finalize(genesisHash)
	if err != nil {
		return err
	}
	go func() {
		for {
			msg, err := h.GetTripodHeader().ChainEnv.P2pNetwork.SubP2P(StartBlockTopic)
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

			ok := h.VerifyBlock(p2pBlock.CompactBlock)
			if !ok {
				logrus.Warnf("p2pBlock(%s) verify failed", p2pBlock.Hash.String())
				continue
			}

			h.recvChan <- p2pBlock
		}
	}()
	return nil
}

func (h *Poa) StartBlock(block *CompactBlock) error {
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
			return nil
		}
	}

	txns, err := h.GetTripodHeader().ChainEnv.Pool.Pack(3000)
	if err != nil {
		return err
	}
	hashes := FromArray(txns...).Hashes()
	block.TxnsHashes = hashes

	txnRoot, err := MakeTxnRoot(txns)
	if err != nil {
		return err
	}
	block.TxnRoot = txnRoot

	byt, _ := block.Encode()
	block.Hash = BytesToHash(Sha256(byt))

	// miner signs block
	block.MinerSignature, err = h.myPrivKey.SignData(block.Hash.Bytes())
	if err != nil {
		return err
	}
	block.MinerPubkey = h.myPubkey.BytesWithType()

	err = h.GetTripodHeader().ChainEnv.Pool.Reset(block)
	if err != nil {
		return err
	}

	h.GetTripodHeader().ChainEnv.State.StartBlock(block.Hash)

	rawBlock := &Block{
		CompactBlock: block,
		Txns:         txns,
	}

	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		return err
	}

	return h.GetTripodHeader().ChainEnv.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
}

func (h *Poa) EndBlock(block *CompactBlock) error {
	chain := h.GetTripodHeader().ChainEnv.Chain

	err := h.GetTripodHeader().ChainEnv.Execute(block)
	if err != nil {
		return err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
		Info("append block")

	h.GetTripodHeader().ChainEnv.State.FinalizeBlock(block.Hash)

	return nil
}

func (h *Poa) FinalizeBlock(block *CompactBlock) error {
	logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
		Info("finalize block")
	return h.GetTripodHeader().ChainEnv.Chain.Finalize(block.Hash)
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

func (h *Poa) useP2pOrSkip(localBlock *CompactBlock) bool {
LOOP:
	select {
	case p2pBlock := <-h.recvChan:
		if h.getCurrentHeight() > p2pBlock.Height {
			goto LOOP
		}
		return h.useP2pBlock(localBlock, p2pBlock)
	case <-time.NewTicker(h.calulateWaitTime(localBlock)).C:
		return false
	}
}

func (h *Poa) useP2pBlock(localBlock *CompactBlock, p2pBlock *Block) bool {
	localBlock.CopyFrom(p2pBlock.CompactBlock)
	err := h.GetTripodHeader().ChainEnv.YuDB.SetTxns(localBlock.Hash, p2pBlock.Txns)
	if err != nil {
		logrus.Errorf("set txns of p2p-block(%s) into base error: %v", p2pBlock.Hash.String(), err)
		return true
	}
	h.GetTripodHeader().ChainEnv.State.StartBlock(localBlock.Hash)
	err = h.GetTripodHeader().ChainEnv.Pool.Reset(localBlock)
	if err != nil {
		logrus.Error("clear txpool error: ", err)
	}
	return true
}

func (h *Poa) calulateWaitTime(block *CompactBlock) time.Duration {
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
