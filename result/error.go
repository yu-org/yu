package result

import (
	"fmt"
	. "yu/common"
	. "yu/utils/codec"
)

type Error interface {
	Error() string
	Encode() ([]byte, error)
}

type TxnError struct {
	Caller     Address
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	ExecName   string
	err        error
}

func (e *TxnError) Error() string {
	return fmt.Sprintf(
		"[Error] Caller(%s) call Tripod(%s) Execution(%s) in Block(%s) on Height(%d): %s",
		e.Caller.String(),
		e.TripodName,
		e.ExecName,
		e.BlockHash,
		e.Height,
		e.err.Error(),
	)
}

func (e *TxnError) Encode() ([]byte, error) {
	return GobEncode(e)
}

type BlockError struct {
	BlockStage string
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	err        error
}

func (e *BlockError) Error() string {
	return fmt.Sprintf(
		"[Error] %s Block(%s) on Height(%d) in Tripod(%s): %s",
		e.BlockStage,
		e.BlockHash,
		e.Height,
		e.TripodName,
		e.err.Error(),
	)
}

func (e *BlockError) Encode() ([]byte, error) {
	return GobEncode(e)
}

type Errors []Error

func ToErrors(errors []Error) Errors {
	var es Errors
	es = append(es, errors...)
	return es
}

func (es Errors) ToArray() []Error {
	return es[:]
}

func (es Errors) Encode() ([]byte, error) {
	return GobEncode(es)
}
