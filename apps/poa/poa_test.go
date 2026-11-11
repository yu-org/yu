package poa

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/kernel"
	. "github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/state/sqlstate"
	"github.com/yu-org/yu/core/tripod"
	"github.com/yu-org/yu/core/txdb"
	"github.com/yu-org/yu/core/txpool"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/infra/p2p"
	"github.com/yu-org/yu/infra/storage/kv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
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
	node1 = newPoa(myPubkey1, myPrivkey1, validators, &PoaConfig{BlockInterval: 3, PackNum: 5000})
	node1.SetChainEnv(&env.ChainEnv{})

	myPubkey2, myPrivkey2, validators = InitDefaultKeypairs(1)
	node2 = newPoa(myPubkey2, myPrivkey2, validators, &PoaConfig{BlockInterval: 3, PackNum: 5000})
	node2.SetChainEnv(&env.ChainEnv{})

	myPubkey3, myPrivkey3, validators = InitDefaultKeypairs(2)
	node3 = newPoa(myPubkey3, myPrivkey3, validators, &PoaConfig{BlockInterval: 3, PackNum: 5000})
	node3.SetChainEnv(&env.ChainEnv{})
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

	assert.NoError(t, node1.VerifyBlock(block))
	assert.NoError(t, node2.VerifyBlock(block))
	assert.NoError(t, node3.VerifyBlock(block))
}

// TestChainNet replays sealed SQL-state blocks on three isolated in-process nodes.
func TestChainNet(t *testing.T) {
	mock := p2p.NewMockP2p(3)
	mock.AddTopic(StartBlockTopic)
	nodes := make([]*Poa, 3)
	kernels := make([]*kernel.Kernel, 3)
	for i := range nodes {
		nodes[i], kernels[i] = newSQLPoaNode(t, i, mock)
	}
	for round := 0; round < 3; round++ {
		producer := round % len(nodes)
		block, err := kernels[producer].LocalRun()
		if err != nil {
			t.Fatal(err)
		}
		if block.StateRoot == NullHash {
			t.Fatal("sealed block lacks state root")
		}
		for i := range nodes {
			encoded, err := mock.SubP2P(StartBlockTopic)
			if err != nil {
				t.Fatal(err)
			}
			if i == producer {
				continue
			}
			received, err := DecodeBlock(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if err := nodes[i].VerifyBlock(received); err != nil {
				t.Fatal(err)
			}
			nodes[i].EndBlock(received)
			nodes[i].FinalizeBlock(received)
			head, err := nodes[i].Chain.LastFinalized()
			if err != nil {
				t.Fatal(err)
			}
			if head.Hash != block.Hash {
				t.Fatal("nodes disagree on finalized block")
			}
		}
	}
}

// newSQLPoaNode builds an isolated node without starting competing P2P listeners.
func newSQLPoaNode(t *testing.T, index int, network *p2p.MockP2p) (*Poa, *kernel.Kernel) {
	t.Helper()
	cfg := config.InitDefaultCfg()
	cfg.BlockChain.ChainDB.Dsn = filepath.Join(t.TempDir(), "chain.db")
	pub, priv, validators := InitDefaultKeypairs(index)
	node := newPoa(pub, priv, validators, &PoaConfig{BlockInterval: 1, PackNum: 10})
	kvdb, err := kv.NewKvdb(&config.KVconf{KvType: "bolt", Path: filepath.Join(t.TempDir(), "tx.db")})
	if err != nil {
		t.Fatal(err)
	}
	txDB, err := txdb.NewTxDB(FullNode, kvdb, config.TxnConf{})
	if err != nil {
		t.Fatal(err)
	}
	chain := blockchain.NewBlockChain(FullNode, &cfg.BlockChain, txDB)
	genesis := &Block{Header: &Header{Hash: BytesToHash(Sha256([]byte("sql-state-genesis")))}}
	if err := chain.SetGenesis(genesis); err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "state.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	store, err := sqlstate.NewState(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitializeState(genesis); err != nil {
		t.Fatal(err)
	}
	chainEnv := &env.ChainEnv{State: store, Chain: chain, TxDB: txDB, Pool: txpool.WithDefaultChecks(FullNode, &cfg.Txpool), P2pNetwork: network}
	k := kernel.NewKernel(cfg, chainEnv, tripod.NewLand()).WithTripods(node)
	return node, k
}
