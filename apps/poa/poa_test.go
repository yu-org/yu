package poa

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockbase"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/chain_env"
	"github.com/yu-org/yu/core/kernel"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	"testing"
)

var (
	myPubkey1, myPubkey2, myPubkey3    PubKey
	myPrivkey1, myPrivkey2, myPrivkey3 PrivKey
	validators                         []ValidatorAddrIp
	node1, node2, node3                *Poa
)

func initGlobalVars() {
	myPubkey1, myPrivkey1, validators = InitKeypair(0)
	node1 = NewPoa(myPubkey1, myPrivkey1, validators)
	println("addr1 = ", myPubkey1.Address().String())

	myPubkey2, myPrivkey2, validators = InitKeypair(1)
	node2 = NewPoa(myPubkey2, myPrivkey2, validators)
	println("addr2 = ", myPubkey2.Address().String())

	myPubkey3, myPrivkey3, validators = InitKeypair(2)
	node3 = NewPoa(myPubkey3, myPrivkey3, validators)
	println("addr3 = ", myPubkey3.Address().String())
}

func TestCompeteLeader(t *testing.T) {
	initGlobalVars()

	for i := 1; i <= 30; i++ {
		bn := BlockNum(i)
		t.Log("block number = ", bn)
		addr1 := node1.CompeteLeader(bn)
		addr2 := node2.CompeteLeader(bn)
		addr3 := node3.CompeteLeader(bn)

		mod := (bn - 1) % 3
		switch mod {
		case 0:
			assert.Equal(t, myPubkey1.Address().String(), addr1.String(), "addr = %s", addr1.String())
		case 1:
			assert.Equal(t, myPubkey2.Address().String(), addr2.String(), "addr = %s", addr2.String())
		case 2:
			assert.Equal(t, myPubkey3.Address().String(), addr3.String(), "addr = %s", addr3.String())
		}
	}
}

func TestVerifyBlock(t *testing.T) {
	initGlobalVars()

	miner := myPubkey1
	signer, err := myPrivkey1.SignData(NullHash.Bytes())
	if err != nil {
		t.Fatal("sign blockhash error: ", err)
	}

	block := &CompactBlock{
		Header: &Header{
			ChainID:        0,
			PrevHash:       NullHash,
			Hash:           NullHash,
			Height:         0,
			TxnRoot:        NullHash,
			StateRoot:      NullHash,
			Timestamp:      0,
			PeerID:         "",
			Extra:          nil,
			LeiLimit:       0,
			LeiUsed:        0,
			MinerPubkey:    miner.BytesWithType(),
			MinerSignature: signer,
			Validators:     nil,
			ProofBlockHash: NullHash,
			ProofHeight:    0,
			Proof:          nil,
			Nonce:          0,
			Difficulty:     0,
		},
		TxnsHashes: nil,
	}

	assert.True(t, node1.VerifyBlock(block))
	assert.True(t, node2.VerifyBlock(block))
	assert.True(t, node3.VerifyBlock(block))
}

func TestChainNet(t *testing.T) {
	initGlobalVars()

	mockP2P := p2p.NewMockP2p()
	go runNode("node1", node1, mockP2P)
	go runNode("node2", node2, mockP2P)
	go runNode("node3", node3, mockP2P)
}

func runNode(cfgPath string, poaNode tripod.Tripod, mockP2P *p2p.MockP2p) {
	cfg := config.InitDefaultCfgWithDir(cfgPath)

	land := tripod.NewLand()
	land.SetTripods(poaNode)

	chain := blockchain.NewBlockChain(&cfg.BlockChain)
	base := blockbase.NewBlockBase(&cfg.BlockBase)
	statedb := state.NewStateDB(&cfg.State)

	env := &chain_env.ChainEnv{
		State:      statedb,
		Chain:      chain,
		Base:       base,
		Pool:       txpool.LocalWithDefaultChecks(&cfg.Txpool, base),
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: mockP2P,
	}
	node1.SetChainEnv(env)

	k := kernel.NewKernel(&cfg, env, land)
	for i := 0; i < 10; i++ {
		err := k.LocalRun()
		if err != nil {
			panic(err)
		}
	}

}
