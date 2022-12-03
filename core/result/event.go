package result

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	. "github.com/yu-org/yu/common"
)

type Event struct {
	Caller      Address  `json:"caller"`
	BlockStage  string   `json:"block_stage"`
	BlockHash   Hash     `json:"block_hash"`
	Height      BlockNum `json:"height"`
	TripodName  string   `json:"tripod_name"`
	WritingName string   `json:"writing_name"`
	Value       []byte   `json:"value"`
	LeiCost     uint64   `json:"lei_cost"`
}

func (e *Event) Hash() (Hash, error) {
	byt, err := e.Encode()
	if err != nil {
		return NullHash, err
	}
	hash := sha256.Sum256(byt)
	return hash, nil
}

func (e *Event) Encode() ([]byte, error) {
	byt, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(EventTypeByt, byt...), nil
}

func (e *Event) Decode(data []byte) error {
	return json.Unmarshal(data[ResultTypeBytesLen:], e)
}

func (e *Event) Type() ResultType {
	return EventType
}

func (e *Event) Sprint() (str string) {
	if e.BlockStage == ExecuteTxnsStage {
		str = fmt.Sprintf(
			"[Event] Caller(%s) call Tripod(%s) Execution(%s) in Block(%s) on Height(%d): %s",
			e.Caller.String(),
			e.TripodName,
			e.WritingName,
			e.BlockHash.String(),
			e.Height,
			e.Value,
		)
	} else {
		str = fmt.Sprintf(
			"[Event] %s Block(%s) on Height(%d) in Tripod(%s): %s",
			e.BlockStage,
			e.BlockHash.String(),
			e.Height,
			e.TripodName,
			e.Value,
		)
	}
	return
}

//type Events []Event
//
//func ToEvents(events []Event) Events {
//	var es Events
//	es = append(es, events...)
//	return es
//}
//
//func (es Events) ToArray() []Event {
//	return es[:]
//}
//
//func (es Events) Encode() ([]byte, error) {
//	return GobEncode(es)
//}
