package poa

import (
	"bytes"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/chain_env"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/context"
	. "github.com/yu-org/yu/keypair"
	. "github.com/yu-org/yu/tripod"
	. "github.com/yu-org/yu/types"
	"time"
)

type Poa struct {
	meta *TripodMeta
	// key: crypto address, generate from pubkey
	validatorsMap map[Address]peer.ID
	myPubkey      PubKey
	myPrivKey     PrivKey

	validatorsList []Address

	env       *ChainEnv
	blockChan chan *Block
	// local node index in addrs
	nodeIdx int
}

func NewPoa(myPubkey PubKey, myPrivkey PrivKey, validatorsMap map[Address]string) *Poa {
	meta := NewTripodMeta("Poa")

	var nodeIdx int

	validatorsAddr := make([]Address, 0)
	validators := make(map[Address]peer.ID)
	for addr, ip := range validatorsMap {
		peerID, err := peer.Decode(ip)
		if err != nil {
			logrus.Fatalf("decode validatorIP(%s) error: %v", ip, err)
		}
		validators[addr] = peerID

		if addr == myPubkey.Address() {
			nodeIdx = len(validatorsAddr)
		}

		validatorsAddr = append(validatorsAddr, addr)
	}

	h := &Poa{
		meta:           meta,
		validatorsMap:  validators,
		validatorsList: validatorsAddr,
		myPubkey:       myPubkey,
		myPrivKey:      myPrivkey,
		blockChan:      make(chan *Block, 10),
		nodeIdx:        nodeIdx,
	}
	h.meta.SetExec(h.JoinValidator, 10000).SetExec(h.QuitValidator, 100)
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

func (h *Poa) GetTripodMeta() *TripodMeta {
	return h.meta
}

func (h *Poa) Name() string {
	return h.meta.Name()
}

func (h *Poa) CheckTxn(txn *SignedTxn) error {
	return nil
}

func (h *Poa) SetChainEnv(env *ChainEnv) {
	h.env = env
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

	chain := h.env.Chain
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
			msg, err := h.env.P2pNetwork.SubP2P(StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			block, err := DecodeBlock(msg)
			if err != nil {
				logrus.Error("decode block from p2p error: ", err)
				continue
			}
			h.blockChan <- block
		}
	}()
	return nil
}

func (h *Poa) StartBlock(block *CompactBlock) error {
	now := time.Now()
	defer func() {
		duration := time.Since(now)
		time.Sleep(3*time.Second - duration)
	}()

	miner := h.CompeteLeader(block.Height)
	logrus.Debugf("compete a leader(%s) in round(%d)", miner, block.Height)
	if miner != h.LocalAddress() {
		if h.useP2pBlock(block) {
			return nil
		}
	}

	txns, err := h.env.Pool.Pack(3000)
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

	h.env.StartBlock(block.Hash)
	err = h.env.Base.SetTxns(block.Hash, txns)
	if err != nil {
		return err
	}

	rawBlock := &Block{
		CompactBlock: block,
		Txns:         txns,
	}

	rawBlockByt, err := rawBlock.Encode()
	if err != nil {
		return err
	}

	return h.env.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
}

func (h *Poa) EndBlock(block *CompactBlock) error {
	chain := h.env.Chain
	pool := h.env.Pool

	err := h.env.Execute(block)
	if err != nil {
		return err
	}

	err = chain.AppendBlock(block)
	if err != nil {
		return err
	}

	logrus.Infof("append block(%d) (%s)", block.Height, block.Hash.String())

	h.env.SetCanRead(block.Hash)

	return pool.Reset()
}

func (h *Poa) FinalizeBlock(block *CompactBlock) error {
	logrus.Infof("Finalize Block(%d) (%s)", block.Height, block.Hash.String())
	return h.env.Chain.Finalize(block.Hash)
}

func (h *Poa) CompeteLeader(blockHeight BlockNum) Address {
	idx := (int(blockHeight) - 1) % len(h.validatorsList)
	return h.validatorsList[idx]
}

func (h *Poa) useP2pBlock(localBlock *CompactBlock) bool {
	var p2pBlock *Block
	select {
	case p2pBlock = <-h.blockChan:
		goto USEP2P
	case <-time.NewTicker(h.calulateWaitTime(localBlock)).C:
		return false
	}
USEP2P:
	logrus.Debugf("accept block(%s), height(%d), miner(%s), signer(%s)",
		p2pBlock.Hash.String(), p2pBlock.Height, ToHex(p2pBlock.MinerPubkey), ToHex(p2pBlock.MinerSignature))
	if localBlock.Height > p2pBlock.Height {
		return false
	}
	if bytes.Equal(p2pBlock.MinerPubkey, h.myPubkey.BytesWithType()) {
		return true
	}
	ok := h.VerifyBlock(p2pBlock.CompactBlock)
	if !ok {
		logrus.Warnf("block(%s) verify failed", p2pBlock.Hash.String())
		return false
	}
	localBlock.CopyFrom(p2pBlock.CompactBlock)
	err := h.env.Base.SetTxns(localBlock.Hash, p2pBlock.Txns)
	if err != nil {
		logrus.Errorf("set txns of p2p-block(%s) into base error: %v", p2pBlock.Hash.String(), err)
		return true
	}
	h.env.StartBlock(localBlock.Hash)
	err = h.env.Pool.RemoveTxns(localBlock.TxnsHashes)
	if err != nil {
		logrus.Error("clear txpool error: ", err)
	}
	return true
}

func (h *Poa) JoinValidator(ctx *context.Context, block *CompactBlock) error {

	return nil
}

func (h *Poa) QuitValidator(ctx *context.Context, block *CompactBlock) error {

	return nil
}

func (h *Poa) calulateWaitTime(block *CompactBlock) time.Duration {
	height := int(block.Height)
	shouldLeaderIdx := (height - 1) % len(h.validatorsList)
	n := shouldLeaderIdx - h.nodeIdx
	if n < 0 {
		n = -n
	}

	return time.Duration(n) * time.Second
}
