package result

import (
	"encoding/json"
	"fmt"
	. "github.com/yu-org/yu/common"
)

type Event struct {
	Caller     Address  `json:"caller"`
	BlockStage string   `json:"block_stage"`
	BlockHash  Hash     `json:"block_hash"`
	Height     BlockNum `json:"height"`
	TripodName string   `json:"tripod_name"`
	ExecName   string   `json:"exec_name"`
	Value      string   `json:"value"`
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
			e.ExecName,
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
