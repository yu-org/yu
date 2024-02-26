package result

import (
	"crypto/sha256"
	"encoding/json"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/trie"
)

type Result struct {
	TxHash      Hash     `json:"tx_hash"`
	Caller      *Address `json:"caller"`
	BlockStage  string   `json:"block_stage"`
	BlockHash   Hash     `json:"block_hash"`
	Height      BlockNum `json:"height"`
	TripodName  string   `json:"tripod_name"`
	WritingName string   `json:"writing_name"`
	LeiCost     uint64   `json:"lei_cost"`

	Events []*Event `json:"events,omitempty"`
	Error  error    `json:"error,omitempty"`
}

func NewResult(events []*Event, err error) *Result {
	return &Result{Events: events, Error: err}
}

func NewWithEvents(events []*Event) *Result {
	return &Result{
		Events: events,
	}
}

func (r *Result) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Result) Decode(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r *Result) Hash() ([]byte, error) {
	byt, err := r.Encode()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(byt)
	return hash[:], err
}

func CaculateReceiptRoot(results []*Result) (Hash, error) {
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
