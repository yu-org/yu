package types

import (
	"crypto/sha256"
	"encoding/json"
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
	Error  error    `json:"error,omitempty"`

	Extra []byte `json:"extra,omitempty"`
}

func NewReceipt(events []*Event, err error) *Receipt {
	return &Receipt{Events: events, Error: err}
}

func NewWithEvents(events []*Event) *Receipt {
	return &Receipt{
		Events: events,
	}
}

func (r *Receipt) FillMetadata(block *Block, stxn *SignedTxn, leiCost uint64) {
	wrCall := stxn.Raw.WrCall

	r.Caller = stxn.GetCallerAddr()
	r.TripodName = wrCall.TripodName
	r.WritingName = wrCall.FuncName
	r.BlockHash = block.Hash
	r.Height = block.Height
	r.LeiCost = leiCost
}

func (r *Receipt) Encode() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Receipt) Decode(data []byte) error {
	return json.Unmarshal(data, r)
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
