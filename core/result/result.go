package result

import (
	"crypto/sha256"
	"encoding/json"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/trie"
)

type Result struct {
	Type int `json:"type"`
	// Event or Error
	Event *Event `json:"event,omitempty"`
	Error *Error `json:"error,omitempty"`
}

func NewEvent(e *Event) *Result {
	return &Result{
		Type:  EventType,
		Event: e,
	}
}

func NewError(e *Error) *Result {
	return &Result{
		Type:  ErrorType,
		Error: e,
	}
}

const (
	EventType = iota
	ErrorType
)

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

func (r *Result) IsEvent() bool {
	return r.Type == EventType
}

func (r *Result) IsError() bool {
	return r.Type == ErrorType
}

func (r *Result) String() string {
	switch r.Type {
	case EventType:
		return r.Event.Sprint()
	case ErrorType:
		return r.Error.Error()
	}
	return "invalid result type"
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
