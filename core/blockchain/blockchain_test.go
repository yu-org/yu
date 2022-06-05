package blockchain

import (
	"github.com/stretchr/testify/assert"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/config"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/codec"
	"testing"
)

var (
	genesisHash = HexToHash("12345")
	block1hash  = HexToHash("1")
	block2hash  = HexToHash("2")
	block3hash  = HexToHash("3")
	uncle1hash  = HexToHash("11")

	genesisBlock = &CompactBlock{
		Header: &Header{
			Hash:   genesisHash,
			Height: 0,
		},
		TxnsHashes: nil,
	}

	block1 = &CompactBlock{
		Header: &Header{
			PrevHash: genesisHash,
			Hash:     block1hash,
			Height:   1,
		},
		TxnsHashes: nil,
	}

	block2 = &CompactBlock{
		Header: &Header{
			PrevHash: block1hash,
			Hash:     block2hash,
			Height:   2,
		},
		TxnsHashes: nil,
	}

	block3 = &CompactBlock{
		Header: &Header{
			PrevHash: block2hash,
			Hash:     block3hash,
			Height:   3,
		},
		TxnsHashes: nil,
	}

	uncleBlock1 = &CompactBlock{
		Header: &Header{
			PrevHash: genesisHash,
			Hash:     uncle1hash,
			Height:   1,
		},
		TxnsHashes: nil,
	}
)

func initChain(t *testing.T) *BlockChain {
	cfg := config.InitDefaultCfgWithDir("test-blockchain")
	codec.GlobalCodec = &codec.RlpCodec{}

	chain := NewBlockChain(&cfg.BlockChain)
	err := chain.SetGenesis(genesisBlock)
	if err != nil {
		t.Fatal(err)
	}
	err = chain.Finalize(genesisHash)
	if err != nil {
		t.Fatal(err)
	}
	return chain
}

func TestEndBlock(t *testing.T) {
	chain := initChain(t)
	insertedBlocks := []*CompactBlock{block1, block2, block3}

	for i, block := range insertedBlocks {
		err := chain.AppendCompactBlock(block)
		if err != nil {
			t.Fatalf("append block(%d) error: %v", i, err)
		}
	}

	block, err := chain.GetEndBlock()
	if err != nil {
		t.Fatal("get end block error: ", err)
	}
	assert.Equal(t, block.Hash, block3.Hash)
}

func TestLastFinalized(t *testing.T) {
	chain := initChain(t)
	insertedBlocks := []*CompactBlock{block1, block2}

	for i, block := range insertedBlocks {
		err := chain.AppendCompactBlock(block)
		if err != nil {
			t.Fatalf("append block(%d) error: %v", i, err)
		}
		err = chain.Finalize(block.Hash)
		if err != nil {
			t.Fatalf("finalize block(%d) error: %v", i, err)
		}
	}
	err := chain.AppendCompactBlock(block3)
	if err != nil {
		t.Fatal("append block3 failed: ", err)
	}

	block, err := chain.LastFinalized()
	if err != nil {
		t.Fatal("get finalized block: ", err)
	}
	assert.Equal(t, block.Hash, block2.Hash)
}

func TestAllBlock(t *testing.T) {
	chain := initChain(t)
	insertedBlocks := []*CompactBlock{block1, block2}
	blockchain := append([]*CompactBlock{genesisBlock}, insertedBlocks...)

	for i, block := range insertedBlocks {
		err := chain.AppendCompactBlock(block)
		if err != nil {
			t.Fatalf("append block(%d) error: %v", i, err)
		}
	}

	blocks, err := chain.GetAllBlocks()
	if err != nil {
		t.Fatal("get all blocks: ", err)
	}
	for i, block := range blocks {
		assert.Equal(t, block.Hash, blockchain[i].Hash)
	}
}

func TestChildrenBlocks(t *testing.T) {
	chain := initChain(t)
	err := chain.AppendCompactBlock(block1)
	if err != nil {
		t.Fatal("append block1 failed: ", err)
	}
	err = chain.AppendCompactBlock(uncleBlock1)
	if err != nil {
		t.Fatal("append uncleBlock1 failed: ", err)
	}
	blocks, err := chain.Children(genesisHash)
	if err != nil {
		t.Fatal("get children blocks failed: ", err)
	}
	children := []*CompactBlock{block1, uncleBlock1}
	for i, block := range blocks {
		assert.Equal(t, block.Hash, children[i].Hash)
	}
}

func TestRangeBlocks(t *testing.T) {
	chain := initChain(t)
	err := chain.AppendCompactBlock(block1)
	if err != nil {
		t.Fatal("append block1 failed: ", err)
	}
	err = chain.AppendCompactBlock(block2)
	if err != nil {
		t.Fatal("append block2 failed: ", err)
	}
	err = chain.AppendCompactBlock(block3)
	if err != nil {
		t.Fatal("append block3 failed: ", err)
	}
	blocks, err := chain.GetRangeBlocks(block1.Height, block3.Height)
	if err != nil {
		t.Fatal("get range blocks failed: ", err)
	}
	assert.Equal(t, len(blocks), 3)
	assert.Equal(t, blocks[0].Hash, block1.Hash)
	assert.Equal(t, blocks[1].Hash, block2.Hash)
	assert.Equal(t, blocks[2].Hash, block3.Hash)
}
