package event

import (
	"bytes"
	"encoding/gob"
)

type Event struct {
}

func (e *Event) Encode() ([]byte, error) {

}

func (e *Event) Print() string {

}

type Events []IEvent

func FromArray(events []IEvent) Events {
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
