package poa

import (
	"bytes"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/MEVless"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/log"
	"go.uber.org/atomic"
	"time"
)

type Poa struct {
	*tripod.Tripod

	MevLess *MEVless.MEVless `tripod:"mevless,omitempty"`

	// key: crypto address, generate from pubkey
	validatorsMap map[common.Address]peer.ID
	myPubkey      keypair.PubKey
	myPrivKey     keypair.PrivKey

	validatorsList []common.Address

	currentHeight *atomic.Uint32

	blockInterval int
	packNum       uint64
	recvChan      chan *types.Block
	// local node index in addrs
	nodeIdx int

	cfg *PoaConfig
}

type ValidatorInfo struct {
	Pubkey keypair.PubKey
	P2pID  peer.ID
}

func NewPoa(cfg *PoaConfig) *Poa {
	pub, priv, infos, err := resolveConfig(cfg)
	if err != nil {
		logrus.Fatal("resolve poa config error: ", err)
	}
	return newPoa(pub, priv, infos, cfg)
}

func newPoa(myPubkey keypair.PubKey, myPrivkey keypair.PrivKey, addrIps []ValidatorInfo, cfg *PoaConfig) *Poa {
	tri := tripod.NewTripod()

	var nodeIdx int

	validatorsAddr := make([]common.Address, 0)
	validators := make(map[common.Address]peer.ID)
	for _, addrIp := range addrIps {
		addr := addrIp.Pubkey.Address()
		validators[addr] = addrIp.P2pID

		if addr == myPubkey.Address() {
			nodeIdx = len(validatorsAddr)
		}

		validatorsAddr = append(validatorsAddr, addr)
	}

	p := &Poa{
		Tripod:         tri,
		validatorsMap:  validators,
		validatorsList: validatorsAddr,
		myPubkey:       myPubkey,
		myPrivKey:      myPrivkey,
		currentHeight:  atomic.NewUint32(0),
		blockInterval:  cfg.BlockInterval,
		packNum:        cfg.PackNum,
		recvChan:       make(chan *types.Block, 10),
		nodeIdx:        nodeIdx,
		cfg:            cfg,
	}
	//p.SetInit(p)
	//p.SetTxnChecker(p)
	//p.SetBlockCycle(p)
	//p.SetBlockVerifier(p)
	return p
}

func (h *Poa) ValidatorsP2pID() (peers []peer.ID) {
	for _, id := range h.validatorsMap {
		peers = append(peers, id)
	}
	return
}

func (h *Poa) LocalAddress() common.Address {
	return h.myPubkey.Address()
}

func (h *Poa) CheckTxn(txn *types.SignedTxn) error {
	// return metamask.CheckMetamaskSig(txn)
	return nil
}

func (h *Poa) VerifyBlock(block *types.Block) error {
	minerPubkey, err := keypair.PubKeyFromBytes(block.MinerPubkey)
	if err != nil {
		logrus.Warnf("parse pubkey(%s) error: %v", block.MinerPubkey, err)
		return err
	}
	if _, ok := h.validatorsMap[minerPubkey.Address()]; !ok {
		logrus.Warn("illegal miner: ", minerPubkey.StringWithType())
		return errors.Errorf("miner(%s) is not validator", minerPubkey.Address())
	}
	if !minerPubkey.VerifySignature(block.Hash.Bytes(), block.MinerSignature) {
		return yerror.BlockSignatureIllegal(block.Hash)
	}
	return nil
}

func (h *Poa) InitChain(block *types.Block) {

	go func() {
		for {
			msg, err := h.P2pNetwork.SubP2P(common.StartBlockTopic)
			if err != nil {
				logrus.Error("subscribe message from P2P error: ", err)
				continue
			}
			p2pBlock, err := types.DecodeBlock(msg)
			if err != nil {
				logrus.Error("decode p2pBlock from p2p error: ", err)
				continue
			}
			if bytes.Equal(p2pBlock.MinerPubkey, h.myPubkey.BytesWithType()) {
				continue
			}

			logrus.Debugf("accept block(%s), height(%d), miner(%s)",
				p2pBlock.Hash.String(), p2pBlock.Height, common.ToHex(p2pBlock.MinerPubkey))

			if h.getCurrentHeight() > p2pBlock.Height {
				continue
			}

			err = h.RangeList(func(tri *tripod.Tripod) error {
				return tri.BlockVerifier.VerifyBlock(block)
			})
			if err != nil {
				logrus.Warnf("p2pBlock(%s) verify failed: %s", p2pBlock.Hash, err)
				continue
			}

			h.recvChan <- p2pBlock
		}
	}()
}

func (h *Poa) StartBlock(block *types.Block) {
	now := time.Now()
	defer func() {
		duration := time.Since(now)
		// fmt.Println("-------start-block last: ", duration.String(), "block-number = ", block.Height)
		time.Sleep(time.Duration(h.blockInterval)*time.Millisecond - duration)
	}()

	h.setCurrentHeight(block.Height)

	if h.cfg.PrettyLog {
		log.StarConsole.Info(fmt.Sprintf("start a new block, height=%d", block.Height))
	}

	if !h.AmILeader(block.Height) {
		if h.useP2pOrSkip(block) {
			logrus.Infof("--------USE P2P Height(%d) block(%s) miner(%s)",
				block.Height, block.Hash.String(), common.ToHex(block.MinerPubkey))
			return
		}
	}

	logrus.Infof(" I am Leader! I mine the block for height (%d)! ", block.Height)
	var (
		txns []*types.SignedTxn
		err  error
	)

	if h.MevLess != nil {
		txns, err = h.MevLess.Pack(block.Height, h.packNum)
	} else {
		txns, err = h.Pool.Pack(h.packNum)
	}

	if err != nil {
		logrus.Panic("pack txns from pool: ", err)
	}

	// logrus.Info("---- the num of pack txns is ", len(txns))

	txnRoot, err := types.MakeTxnRoot(txns)
	if err != nil {
		logrus.Panic("make txn-root failed: ", err)
	}
	block.TxnRoot = txnRoot

	byt, _ := block.Encode()
	block.Hash = common.BytesToHash(common.Sha256(byt))

	// miner signs block
	block.MinerSignature, err = h.myPrivKey.SignData(block.Hash.Bytes())
	if err != nil {
		logrus.Panic("sign block failed: ", err)
	}
	block.MinerPubkey = h.myPubkey.BytesWithType()

	block.SetTxns(txns)

	h.State.StartBlock(block)

	blockByt, err := block.Encode()
	if err != nil {
		logrus.Panic("encode raw-block failed: ", err)
	}

	err = h.P2pNetwork.PubP2P(common.StartBlockTopic, blockByt)
	if err != nil {
		logrus.Panic("publish block to p2p failed: ", err)
	}
}

func (h *Poa) EndBlock(block *types.Block) {
	chain := h.Chain

	// now := time.Now()
	err := h.Execute(block)
	if err != nil {
		logrus.Panic("execute block failed: ", err)
	}
	// TODO: sync the state (execute receipt) with other nodes

	err = chain.AppendBlock(block)
	if err != nil {
		logrus.Panic("append block failed: ", err)
	}
	// fmt.Println("execute block last: ", time.Since(now).String())

	err = h.Pool.Reset(block.Txns)
	if err != nil {
		logrus.Panic("reset pool failed: ", err)
	}

	// log.PlusLog().Info(fmt.Sprintf("append block, height=%d, hash=%s", block.Height, block.Hash.String()))

	//logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
	//	Info("append block")

	h.State.FinalizeBlock(block)
}

func (h *Poa) FinalizeBlock(block *types.Block) {
	//logrus.WithField("block-height", block.Height).WithField("block-hash", block.Hash.String()).
	//	Info("finalize block")

	if h.cfg.PrettyLog {
		log.DoubleLineConsole.Info(fmt.Sprintf("finalize block, height=%d, hash=%s", block.Height, block.Hash.String()))
	}
	h.Chain.Finalize(block)
}

func (h *Poa) CompeteLeader(blockHeight common.BlockNum) common.Address {
	idx := (int(blockHeight) - 1) % len(h.validatorsList)
	leader := h.validatorsList[idx]
	logrus.Debugf("compete a leader(%s) in round(%d)", leader.String(), blockHeight)
	return leader
}

func (h *Poa) AmILeader(blockHeight common.BlockNum) bool {
	return h.CompeteLeader(blockHeight) == h.LocalAddress()
}

func (h *Poa) IsValidator(addr common.Address) bool {
	_, ok := h.validatorsMap[addr]
	return ok
}

func (h *Poa) useP2pOrSkip(localBlock *types.Block) bool {
LOOP:
	select {
	case p2pBlock := <-h.recvChan:
		if h.getCurrentHeight() > p2pBlock.Height {
			goto LOOP
		}
		localBlock.CopyFrom(p2pBlock)
		h.State.StartBlock(localBlock)
		return true
	case <-time.NewTicker(h.calculateWaitTime(localBlock)).C:
		return false
	}
}

func (h *Poa) calculateWaitTime(block *types.Block) time.Duration {
	//height := int(block.Height)
	//shouldLeaderIdx := (height - 1) % len(h.validatorsList)
	//n := shouldLeaderIdx - h.nodeIdx
	//if n < 0 {
	//	n = -n
	//}

	return time.Duration(h.blockInterval) * time.Millisecond
}

func (h *Poa) getCurrentHeight() common.BlockNum {
	return common.BlockNum(h.currentHeight.Load())
}

func (h *Poa) setCurrentHeight(height common.BlockNum) {
	h.currentHeight.Store(uint32(height))
}
