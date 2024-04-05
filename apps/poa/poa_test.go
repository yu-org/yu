package poa

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/kernel"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/state"
	"github.com/yu-org/yu/core/subscribe"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txdb"
	"github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/infra/storage/kv"
	"sync"
	"testing"
)

var (
	myPubkey1, myPubkey2, myPubkey3    PubKey
	myPrivkey1, myPrivkey2, myPrivkey3 PrivKey
	validators                         []ValidatorInfo
	node1, node2, node3                *Poa
)

func initGlobalVars() {
	myPubkey1, myPrivkey1, validators = InitDefaultKeypairs(0)
	node1 = newPoa(myPubkey1, myPrivkey1, validators, 3, 5000)
	println("addr1 = ", myPubkey1.Address().String())

	myPubkey2, myPrivkey2, validators = InitDefaultKeypairs(1)
	node2 = newPoa(myPubkey2, myPrivkey2, validators, 3, 5000)
	println("addr2 = ", myPubkey2.Address().String())

	myPubkey3, myPrivkey3, validators = InitDefaultKeypairs(2)
	node3 = newPoa(myPubkey3, myPrivkey3, validators, 3, 5000)
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

	cblock := &CompactBlock{
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
	block := &Block{
		Header: cblock.Header,
		Txns:   nil,
	}

	assert.True(t, node1.VerifyBlock(block))
	assert.True(t, node2.VerifyBlock(block))
	assert.True(t, node3.VerifyBlock(block))
}

func TestChainNet(t *testing.T) {
	initGlobalVars()

	mockP2P := p2p.NewMockP2p(3)
	mockP2P.AddTopic(StartBlockTopic)
	mockP2P.AddTopic(EndBlockTopic)
	mockP2P.AddTopic(FinalizeBlockTopic)

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go runNode(node1, mockP2P, wg)
	go runNode(node2, mockP2P, wg)
	go runNode(node3, mockP2P, wg)
	wg.Wait()
}

func runNode(poaNode *Poa, mockP2P *p2p.MockP2p, wg *sync.WaitGroup) {
	cfg := config.InitDefaultCfg()

	land := tripod.NewLand()
	land.SetTripods(poaNode.Tripod)

	kvdb, err := kv.NewKvdb(&config.KVconf{
		KvType: "bolt",
		Path:   "test_poa",
	})
	if err != nil {
		panic("init kvdb error: " + err.Error())
	}

	base := txdb.NewTxDB(FullNode, kvdb)
	chain := blockchain.NewBlockChain(FullNode, &cfg.BlockChain, base)
	statedb := state.NewStateDB(kvdb)

	env := &env.ChainEnv{
		State:      statedb,
		Chain:      chain,
		TxDB:       base,
		Pool:       txpool.WithDefaultChecks(FullNode, &cfg.Txpool, base),
		Sub:        subscribe.NewSubscription(),
		P2pNetwork: mockP2P,
	}
	poaNode.SetChainEnv(env)

	k := kernel.NewKernel(cfg, env, land)
	for i := 0; i < 10; i++ {
		err := k.LocalRun()
		if err != nil {
			panic(err)
		}
	}

	wg.Done()
}
