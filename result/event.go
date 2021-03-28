package result

import (
	"fmt"
	. "yu/common"
	. "yu/utils/codec"
)

type Event struct {
	Caller     Address
	BlockStage string
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	ExecName   string
	Value      Codec
}

type Codec interface {
	Encode() ([]byte, error)
	Decode(data []byte) error
}

func (e *Event) Encode() ([]byte, error) {
	return GobEncode(e)
}

func (e *Event) ValueStr() (string, error) {
	byt, err := e.Value.Encode()
	if err != nil {
		return "", err
	}
	return ToHex(byt), nil
}

func (e *Event) Sprint() (str string) {
	if e.BlockStage == ExecuteTxnsStage {
		str = fmt.Sprintf(
			"[Event] Caller(%s) call Tripod(%s) Execution(%s) in Block(%s) on Height(%d): %v",
			e.Caller.String(),
			e.TripodName,
			e.ExecName,
			e.BlockHash,
			e.Height,
			e.Value,
		)
	} else {
		str = fmt.Sprintf(
			"[Event] %s Block(%s) on Height(%d) in Tripod(%s): %v",
			e.BlockStage,
			e.BlockHash,
			e.Height,
			e.TripodName,
			e.Value,
		)
	}
	return
}

type Events []Event

func ToEvents(events []Event) Events {
	var es Events
	es = append(es, events...)
	return es
}

func (es Events) ToArray() []Event {
	return es[:]
}

func (es Events) Encode() ([]byte, error) {
	return GobEncode(es)
}
