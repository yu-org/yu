package result

import (
	"fmt"
	. "yu/common"
	. "yu/utils/codec"
)

type Error struct {
	Caller     Address
	BlockStage string
	BlockHash  Hash
	Height     BlockNum
	TripodName string
	ExecName   string
	Err        error
}

func (e *Error) Type() ResultType {
	return ErrorType
}

func (e *Error) Error() (str string) {
	if e.BlockStage == ExecuteTxnsStage {
		str = fmt.Sprintf(
			"[Error] Caller(%s) call Tripod(%s) Execution(%s) in Block(%s) on Height(%d): %s",
			e.Caller.String(),
			e.TripodName,
			e.ExecName,
			e.BlockHash,
			e.Height,
			e.Err.Error(),
		)
	} else {
		str = fmt.Sprintf(
			"[Error] %s Block(%s) on Height(%d) in Tripod(%s): %s",
			e.BlockStage,
			e.BlockHash,
			e.Height,
			e.TripodName,
			e.Err.Error(),
		)
	}
	return
}

func (e *Error) Encode() ([]byte, error) {
	return GlobalCodec.EncodeToBytes(e)
}

func (e *Error) Decode(data []byte) error {
	return GlobalCodec.DecodeBytes(data, e)
}

//
//type Errors []Error
//
//func ToErrors(errors []Error) Errors {
//	var es Errors
//	es = append(es, errors...)
//	return es
//}
//
//func (es Errors) ToArray() []Error {
//	return es[:]
//}
//
//func (es Errors) Encode() ([]byte, error) {
//	return GobEncode(es)
//}
