package result

import (
	"errors"
	. "github.com/yu-org/yu/common"
	"github.com/yu-org/yu/infra/trie"
	"strconv"
)

type Result interface {
	Type() ResultType
	Encode() ([]byte, error)
	Decode(data []byte) error
}

type ResultType int

const (
	EventType ResultType = iota
	ErrorType

	ResultTypeBytesLen = 1
)

var (
	EventTypeByt = []byte(strconv.Itoa(int(EventType)))
	ErrorTypeByt = []byte(strconv.Itoa(int(ErrorType)))
)

// this func use for clients
func DecodeResult(data []byte) (Result, error) {
	resultTypeByt := data[:ResultTypeBytesLen]

	resultType, err := strconv.Atoi(string(resultTypeByt))
	if err != nil {
		return nil, err
	}

	switch ResultType(resultType) {
	case EventType:
		event := &Event{}
		err := event.Decode(data)
		return event, err
	case ErrorType:
		er := &Error{}
		err := er.Decode(data)
		return er, err
	}
	return nil, errors.New("no result type")
}

func CaculateReceiptRoot(results []Result) (Hash, error) {
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
