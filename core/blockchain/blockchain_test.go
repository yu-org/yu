package blockchain

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
)

var (
	genesisHash = HexToHash("12345")
	block1hash  = HexToHash("1")
	block2hash  = HexToHash("2")
	block3hash  = HexToHash("3")
	uncle1hash  = HexToHash("11")

	genesisBlock = &Block{
		Header: &Header{
			Hash:   genesisHash,
			Height: 0,
		},
	}

	block1 = &Block{
		Header: &Header{
			PrevHash: genesisHash,
			Hash:     block1hash,
			Height:   1,
		},
	}

	block2 = &Block{
		Header: &Header{
			PrevHash: block1hash,
			Hash:     block2hash,
			Height:   2,
		},
	}

	block3 = &Block{
		Header: &Header{
			PrevHash: block2hash,
			Hash:     block3hash,
			Height:   3,
		},
	}

	uncleBlock1 = &Block{
		Header: &Header{
			PrevHash: genesisHash,
			Hash:     uncle1hash,
			Height:   1,
		},
	}
)

// memTxDB is an ItxDB doing nothing, the blockchain tests below only use empty blocks.
type memTxDB struct{}

func (memTxDB) GetTxn(Hash) (*SignedTxn, error)        { return nil, nil }
func (memTxDB) GetTxns([]Hash) ([]*SignedTxn, error)   { return nil, nil }
func (memTxDB) ExistTxn(Hash) bool                     { return false }
func (memTxDB) SetTxns([]*SignedTxn) error             { return nil }
func (memTxDB) SetReceipts(map[Hash]*Receipt) error    { return nil }
func (memTxDB) GetReceipt(Hash) (*Receipt, error)      { return nil, nil }
func (memTxDB) GetReceipts([]Hash) ([]*Receipt, error) { return nil, nil }
func (memTxDB) SetReceipt(Hash, *Receipt) error        { return nil }

func initChain(t *testing.T) *BlockChain {
	cfg := config.InitDefaultCfg()
	cfg.BlockChain.ChainDB.Dsn = filepath.Join(t.TempDir(), "chain.db")

	chain := NewBlockChain(FullNode, &cfg.BlockChain, memTxDB{})
	err := chain.SetGenesis(genesisBlock)
	if err != nil {
		t.Fatal(err)
	}
	err = chain.Finalize(genesisBlock)
	if err != nil {
		t.Fatal(err)
	}
	return chain
}

func appendBlocks(t *testing.T, chain *BlockChain, blocks ...*Block) {
	t.Helper()
	for i, block := range blocks {
		if err := chain.AppendBlock(block); err != nil {
			t.Fatalf("append block(%d) error: %v", i, err)
		}
	}
}

func TestEndBlock(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)

	block, err := chain.GetEndBlock()
	if err != nil {
		t.Fatal("get end block error: ", err)
	}
	assert.Equal(t, block3.Hash, block.Hash)
}

func TestLastFinalized(t *testing.T) {
	chain := initChain(t)

	for i, block := range []*Block{block1, block2} {
		if err := chain.AppendBlock(block); err != nil {
			t.Fatalf("append block(%d) error: %v", i, err)
		}
		if err := chain.Finalize(block); err != nil {
			t.Fatalf("finalize block(%d) error: %v", i, err)
		}
	}
	appendBlocks(t, chain, block3)

	block, err := chain.LastFinalized()
	if err != nil {
		t.Fatal("get finalized block: ", err)
	}
	assert.Equal(t, block2.Hash, block.Hash)
}

func TestAllBlock(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2)

	blocks, err := chain.GetAllCompactBlocks()
	if err != nil {
		t.Fatal("get all blocks: ", err)
	}
	blockchain := []*Block{genesisBlock, block1, block2}
	assert.Equal(t, len(blockchain), len(blocks))
	for i, block := range blocks {
		assert.Equal(t, blockchain[i].Hash, block.Hash)
	}
}

func TestChildrenBlocks(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, uncleBlock1)

	blocks, err := chain.ChildrenCompact(genesisHash)
	if err != nil {
		t.Fatal("get children blocks failed: ", err)
	}
	children := []*Block{block1, uncleBlock1}
	assert.Equal(t, len(children), len(blocks))
	for i, block := range blocks {
		assert.Equal(t, children[i].Hash, block.Hash)
	}
}

func TestRangeBlocks(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)

	blocks, err := chain.GetRangeBlocks(block1.Height, block3.Height)
	if err != nil {
		t.Fatal("get range blocks failed: ", err)
	}
	assert.Equal(t, 3, len(blocks))
	assert.Equal(t, block1.Hash, blocks[0].Hash)
	assert.Equal(t, block2.Hash, blocks[1].Hash)
	assert.Equal(t, block3.Hash, blocks[2].Hash)
}

// heights returns the heights left in the chain DB, sorted ascending.
func heights(t *testing.T, chain *BlockChain) []BlockNum {
	t.Helper()
	blocks, err := chain.GetAllCompactBlocks()
	if err != nil {
		t.Fatal("get all blocks failed: ", err)
	}
	hs := make([]BlockNum, 0, len(blocks))
	for _, block := range blocks {
		hs = append(hs, block.Height)
	}
	return hs
}

func TestPruneAll(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)
	if err := chain.Finalize(block1); err != nil {
		t.Fatal("finalize block1 failed: ", err)
	}

	if err := chain.PruneAll(); err != nil {
		t.Fatal("prune all failed: ", err)
	}

	// Only the finalized genesis and block1 are left.
	assert.Equal(t, []BlockNum{0, 1}, heights(t, chain))

	// The head falls back to the highest block still in the DB.
	end, err := chain.GetEndBlock()
	if err != nil {
		t.Fatal("get end block failed: ", err)
	}
	assert.Equal(t, block1.Hash, end.Hash)

	_, err = chain.GetBlockByHeight(block2.Height)
	assert.ErrorIs(t, err, yerror.ErrBlockNotFound)
}

func TestPruneAfter(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)

	if err := chain.PruneAfter(block2.Height); err != nil {
		t.Fatal("prune after failed: ", err)
	}

	// `height` is included, block1 stays untouched.
	assert.Equal(t, []BlockNum{0, 1}, heights(t, chain))
}

func TestPruneAfterKeepsFinalized(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)
	if err := chain.Finalize(block3); err != nil {
		t.Fatal("finalize block3 failed: ", err)
	}

	if err := chain.PruneAfter(block1.Height); err != nil {
		t.Fatal("prune after failed: ", err)
	}

	assert.Equal(t, []BlockNum{0, 3}, heights(t, chain))
}

func TestPrune(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, block2, block3)
	if err := chain.Finalize(block1); err != nil {
		t.Fatal("finalize block1 failed: ", err)
	}

	if err := chain.Prune(); err != nil {
		t.Fatal("prune failed: ", err)
	}

	assert.Equal(t, []BlockNum{0, 1}, heights(t, chain))
}

func TestPruneUncleBlocks(t *testing.T) {
	chain := initChain(t)
	appendBlocks(t, chain, block1, uncleBlock1, block2)
	if err := chain.Finalize(block1); err != nil {
		t.Fatal("finalize block1 failed: ", err)
	}

	// The uncle sits at the last finalized height, Prune() only touches the blocks above it.
	if err := chain.Prune(); err != nil {
		t.Fatal("prune failed: ", err)
	}
	assert.Equal(t, []BlockNum{0, 1, 1}, heights(t, chain))

	// PruneAfter reaches it, the finalized block1 stays.
	if err := chain.PruneAfter(block1.Height); err != nil {
		t.Fatal("prune after failed: ", err)
	}
	blocks, err := chain.GetAllCompactBlocks()
	if err != nil {
		t.Fatal("get all blocks failed: ", err)
	}
	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, block1.Hash, blocks[1].Hash)
}

func TestPruneWithoutFinalizedBlock(t *testing.T) {
	cfg := config.InitDefaultCfg()
	cfg.BlockChain.ChainDB.Dsn = filepath.Join(t.TempDir(), "chain.db")
	chain := NewBlockChain(FullNode, &cfg.BlockChain, memTxDB{})

	if err := chain.SetGenesis(genesisBlock); err != nil {
		t.Fatal(err)
	}
	appendBlocks(t, chain, block1)

	assert.ErrorIs(t, chain.Prune(), yerror.ErrBlockNotFound)

	// Nothing is finalized, so PruneAll wipes the whole chain, genesis included.
	if err := chain.PruneAll(); err != nil {
		t.Fatal("prune all failed: ", err)
	}
	assert.Empty(t, heights(t, chain))
}
