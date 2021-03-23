package result

import (
	"fmt"
	. "yu/common"
	. "yu/utils/codec"
)

type Event interface {
	String() string
	Encode() ([]byte, error)
}

type TxnEvent struct {
	Caller     Address
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	ExecName   string
	Value      interface{}
}

func (e *TxnEvent) Encode() ([]byte, error) {
	return GobEncode(e)
}

func (e *TxnEvent) String() string {
	return fmt.Sprintf(
		"[Event] Caller(%s) call Tripod(%s) Execution(%s) in Block(%s) on Height(%d): %v",
		e.Caller.String(),
		e.TripodName,
		e.ExecName,
		e.BlockHash,
		e.Height,
		e.Value,
	)
}

type BlockEvent struct {
	BlockStage string
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	Value      interface{}
}

func (e *BlockEvent) Encode() ([]byte, error) {
	return GobEncode(e)
}

func (e *BlockEvent) String() string {
	return fmt.Sprintf("[Event] %s Block(%s) on Height(%d) in Tripod(%s): %v",
		e.BlockStage,
		e.BlockHash,
		e.Height,
		e.TripodName,
		e.Value,
	)
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
