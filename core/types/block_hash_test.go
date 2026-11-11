package types

import (
	"testing"

	"github.com/yu-org/yu/common"
)

// TestExecutionReceiptsIgnoreLocalMetadata checks replay-independent receipt commitments.
func TestExecutionReceiptsIgnoreLocalMetadata(t *testing.T) {
	a, b := common.HexToHash("01"), common.HexToHash("02")
	receipts := map[common.Hash]*Receipt{a: {TxHash: a, SaveTimestamp: 1}, b: {TxHash: b, SaveTimestamp: 2}}
	root, err := CalculateExecutionReceiptRoot(receipts)
	if err != nil {
		t.Fatal(err)
	}
	receipts[a].SaveTimestamp, receipts[b].SaveTimestamp = 100, 200
	receipts[a].BlockHash = common.HexToHash("ff")
	other, err := CalculateExecutionReceiptRoot(map[common.Hash]*Receipt{b: receipts[b], a: receipts[a]})
	if err != nil {
		t.Fatal(err)
	}
	if root != other {
		t.Fatal("receipt root depends on map order or local metadata")
	}
	receipts[a].Error = "failed"
	changed, err := CalculateExecutionReceiptRoot(receipts)
	if err != nil {
		t.Fatal(err)
	}
	if root == changed {
		t.Fatal("receipt root does not commit to execution results")
	}
}

// TestHeaderHashCommitsToStateRoot checks sealing and subsequent root tampering.
func TestHeaderHashCommitsToStateRoot(t *testing.T) {
	block := &Block{Header: &Header{StateRoot: common.HexToHash("01")}}
	var err error
	block.TxnRoot, err = MakeTxnRoot(nil)
	if err != nil {
		t.Fatal(err)
	}
	block.Hash, err = ComputeHeaderHash(block)
	if err != nil {
		t.Fatal(err)
	}
	block.MinerSignature = []byte("signature")
	if err := VerifyHeaderCommitments(block); err != nil {
		t.Fatal(err)
	}
	block.StateRoot = common.HexToHash("02")
	if VerifyHeaderCommitments(block) == nil {
		t.Fatal("state root tampering went undetected")
	}
}
