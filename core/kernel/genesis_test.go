package kernel

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/blockchain"
	"github.com/yu-org/yu/core/env"
	"github.com/yu-org/yu/core/tripod"
	. "github.com/yu-org/yu/core/types"
)

// memTxDB is an ItxDB doing nothing, the genesis blocks below have no txns.
type memTxDB struct{}

func (memTxDB) GetTxn(Hash) (*SignedTxn, error)        { return nil, nil }
func (memTxDB) GetTxns([]Hash) ([]*SignedTxn, error)   { return nil, nil }
func (memTxDB) ExistTxn(Hash) bool                     { return false }
func (memTxDB) SetTxns([]*SignedTxn) error             { return nil }
func (memTxDB) SetReceipts(map[Hash]*Receipt) error    { return nil }
func (memTxDB) GetReceipt(Hash) (*Receipt, error)      { return nil, nil }
func (memTxDB) GetReceipts([]Hash) ([]*Receipt, error) { return nil, nil }
func (memTxDB) SetReceipt(Hash, *Receipt) error        { return nil }

type genesisTripod struct {
	*tripod.Tripod
	genesis *Block
	calls   int
}

func (g *genesisTripod) DefineGenesis() *Block {
	g.calls++
	return g.genesis
}

func newGenesisTripod(name string, genesis *Block) *genesisTripod {
	g := &genesisTripod{Tripod: tripod.NewTripodWithName(name), genesis: genesis}
	g.SetInstance(g)
	return g
}

func newGenesisKernel(t *testing.T, tripods ...*genesisTripod) *Kernel {
	cfg := config.InitDefaultCfg()
	cfg.BlockChain.ChainDB.Dsn = filepath.Join(t.TempDir(), "chain.db")
	chain := blockchain.NewBlockChain(FullNode, &cfg.BlockChain, memTxDB{})

	land := tripod.NewLand()
	for _, tri := range tripods {
		land.RegisterTripods(tri.Tripod)
	}
	return &Kernel{ChainEnv: &env.ChainEnv{Chain: chain}, Land: land}
}

func assertFinalizedGenesis(t *testing.T, k *Kernel, hash Hash) {
	t.Helper()
	genesis, err := k.Chain.GetGenesis()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hash, genesis.Hash)

	finalized, err := k.Chain.GetFinalizedCompactBlockByHeight(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, hash, finalized.Hash)
}

func TestInitGenesisDefault(t *testing.T) {
	k := newGenesisKernel(t)
	if err := k.initGenesis(); err != nil {
		t.Fatal(err)
	}
	genesis, err := k.Chain.GetGenesis()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, NullHash, genesis.Hash)
	assertFinalizedGenesis(t, k, genesis.Hash)

	// Every node of the same chain defines the same default genesis.
	other := newGenesisKernel(t)
	if err := other.initGenesis(); err != nil {
		t.Fatal(err)
	}
	assertFinalizedGenesis(t, other, genesis.Hash)
}

func TestInitGenesisFromDefiner(t *testing.T) {
	hash := HexToHash("abcdef")
	definer := newGenesisTripod("definer", &Block{Header: &Header{Hash: hash}})
	k := newGenesisKernel(t, definer)

	if err := k.initGenesis(); err != nil {
		t.Fatal(err)
	}
	assertFinalizedGenesis(t, k, hash)
	assert.Equal(t, 1, definer.calls)
}

func TestInitGenesisKeepsExisting(t *testing.T) {
	hash := HexToHash("abcdef")
	definer := newGenesisTripod("definer", &Block{Header: &Header{Hash: hash}})
	k := newGenesisKernel(t, definer)
	if err := k.initGenesis(); err != nil {
		t.Fatal(err)
	}

	// A restart must not redefine the genesis.
	definer.genesis = &Block{Header: &Header{Hash: HexToHash("123456")}}
	if err := k.initGenesis(); err != nil {
		t.Fatal(err)
	}
	assertFinalizedGenesis(t, k, hash)
	assert.Equal(t, 1, definer.calls)
}

func TestInitGenesisRejectsSeveralDefiners(t *testing.T) {
	k := newGenesisKernel(t,
		newGenesisTripod("definer1", &Block{Header: &Header{Hash: HexToHash("1")}}),
		newGenesisTripod("definer2", &Block{Header: &Header{Hash: HexToHash("2")}}),
	)
	assert.Error(t, k.initGenesis())

	_, err := k.Chain.GetGenesis()
	assert.ErrorIs(t, err, yerror.ErrBlockNotFound)
}

func TestInitGenesisRejectsNilGenesis(t *testing.T) {
	k := newGenesisKernel(t, newGenesisTripod("definer", nil))
	assert.ErrorIs(t, k.initGenesis(), yerror.GenesisBlockIllegal)
}

func TestInitGenesisRejectsNullHash(t *testing.T) {
	k := newGenesisKernel(t, newGenesisTripod("definer", &Block{Header: &Header{}}))
	assert.ErrorIs(t, k.initGenesis(), yerror.GenesisBlockIllegal)

	_, err := k.Chain.GetGenesis()
	assert.ErrorIs(t, err, yerror.ErrBlockNotFound)
}
