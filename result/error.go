package result

import (
	"bytes"
	"encoding/gob"
)

type IError interface {
	Error() string
	Encode() ([]byte, error)
}

type Error struct {
}

func (e *Error) Error() string {

}

func (e *Error) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(e)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Errors []IError

func ToErrors(errors []IError) Errors {
	var es Errors
	es = append(es, errors...)
	return es
}

func (es Errors) ToArray() []IError {
	return es[:]
}

func (es Errors) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(es)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeErrors(data []byte) (Errors, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var es Errors
	err := decoder.Decode(&es)
	if err != nil {
		return nil, err
	}
	return es, nil
}
