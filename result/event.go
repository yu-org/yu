package result

import (
	"bytes"
	"encoding/gob"
	"fmt"
	. "yu/common"
)

type IEvent interface {
	String() string
	Encode() ([]byte, error)
}

type TxnEvent struct {
	Caller     Address
	TripodName string
	ExecName   string
	Value      interface{}
}

func (e *TxnEvent) Encode() ([]byte, error) {
	return gobEncode(e)
}

func (e *TxnEvent) String() string {
	return fmt.Sprintf(
		"[Event] Caller(%s) call Tripod(%s) Execution(%s): %v",
		e.Caller.String(),
		e.TripodName,
		e.ExecName,
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
	return gobEncode(e)
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

type Events []IEvent

func ToEvents(events []IEvent) Events {
	var es Events
	es = append(es, events...)
	return es
}

func (es Events) ToArray() []IEvent {
	return es[:]
}

func (es Events) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(es)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeEvents(data []byte) (Events, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var es Events
	err := decoder.Decode(&es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

func gobEncode(e interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(e)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
