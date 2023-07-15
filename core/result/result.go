package result

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/trie"
)

type Result struct {
	Type int `json:"type"`
	// event or error
	Object any `json:"object"`
}

func NewEvent(e *Event) *Result {
	return &Result{
		Type:   EventType,
		Object: e,
	}
}

func NewError(e *Error) *Result {
	return &Result{
		Type:   ErrorType,
		Object: e,
	}
}

const (
	EventType = iota
	ErrorType
)

func (r *Result) Event() *Event {
	return r.Object.(*Event)
}

func (r *Result) Error() *Error {
	return r.Object.(*Error)
}

func (r *Result) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, r)
	if err != nil {
		return err
	}

	switch r.Type {
	case EventType:
		event := new(Event)
		err = mapstructure.Decode(r.Object, event)
		if err != nil {
			return err
		}
		r.Object = event
	case ErrorType:
		erro := new(Error)
		err = mapstructure.Decode(r.Object, erro)
		if err != nil {
			return err
		}
		r.Object = erro
	}

	return nil
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

func (r *Result) IsEvent() bool {
	return r.Type == EventType
}

func (r *Result) IsError() bool {
	return r.Type == ErrorType
}

func (r *Result) String() string {
	switch r.Type {
	case EventType:
		return r.Object.(*Event).Sprint()
	case ErrorType:
		return r.Object.(*Error).Error()
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
