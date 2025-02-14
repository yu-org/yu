package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/trie"
)

type Receipt struct {
	TxHash      Hash     `json:"tx_hash"`
	Caller      *Address `json:"caller"`
	BlockStage  string   `json:"block_stage"`
	BlockHash   Hash     `json:"block_hash"`
	Height      BlockNum `json:"height"`
	TripodName  string   `json:"tripod_name"`
	WritingName string   `json:"writing_name"`
	LeiCost     uint64   `json:"lei_cost"`

	Events []*Event `json:"events,omitempty"`
	Error  string   `json:"error,omitempty"`

	Extra         []byte `json:"extra,omitempty"`
	SaveTimestamp int64  `json:"save_timestamp,omitempty"`
}

func NewReceipt(events []*Event, err error, extra []byte) *Receipt {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	return &Receipt{Events: events, Error: errStr, Extra: extra}
}

func (r *Receipt) String() string {
	return fmt.Sprintf(
		"Receipt:{ tx_hash: %s, caller: %s, block_stage: %s, block_hash: %s, Height: %d, tripod_name: %s, writing_name: %s, lei_cost: %d, events: %v, error: %v, extra: %s }",
		r.TxHash.String(), r.Caller.String(), r.BlockStage, r.BlockHash.String(), r.Height, r.TripodName, r.WritingName, r.LeiCost, r.Events, r.Error, string(r.Extra))
}

func (r *Receipt) FillMetadata(block *Block, stxn *SignedTxn, leiCost uint64) {
	wrCall := stxn.Raw.WrCall

	r.TxHash = stxn.TxnHash
	r.Caller = stxn.GetCaller()
	r.TripodName = wrCall.TripodName
	r.WritingName = wrCall.FuncName
	r.BlockHash = block.Hash
	r.Height = block.Height
	r.LeiCost = leiCost
	r.SaveTimestamp = time.Now().Unix()
}

func (r *Receipt) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Receipt) Decode(data []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	return decoder.Decode(r)
}

func (r *Receipt) Hash() ([]byte, error) {
	byt, err := r.Encode()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(byt)
	return hash[:], err
}

func CaculateReceiptRoot(results map[Hash]*Receipt) (Hash, error) {
	var receiptsByt []Hash
	for _, result := range results {
		receipt, err := result.Encode()
		if err != nil {
			return NullHash, err
		}
		receiptsByt = append(receiptsByt, BytesToHash(receipt))
	}
	mTree := trie.NewMerkleTree(receiptsByt)
	return mTree.RootNode.Data, nil
}
