package result

import "github.com/sirupsen/logrus"

type Result interface {
	Type() ResultType
	Encode() ([]byte, error)
	Decode(data []byte) error
}

type ResultType int

const (
	EventType ResultType = iota
	ErrorType
)

// this func use for clients
// NOT good implementation
func DecodeResult(data []byte) Result {
	tryEvent := &Event{}
	err := tryEvent.Decode(data)
	if err == nil {
		return tryEvent
	}
	tryError := &Error{}
	err = tryError.Decode(data)
	if err != nil {
		logrus.Panicf("decode result error: %s", err.Error())
	}
	return tryError
}
