package result

import (
	"fmt"
	. "github.com/yu-org/yu/common"
)

type Error struct {
	Caller      *Address `json:"caller"`
	BlockStage  string   `json:"block_stage"`
	BlockHash   Hash     `json:"block_hash"`
	Height      BlockNum `json:"height"`
	TripodName  string   `json:"tripod_name"`
	WritingName string   `json:"writing_name"`
	Err         string   `json:"err"`
}

func (e *Error) Error() (str string) {
	if e.BlockStage == ExecuteTxnsStage {
		str = fmt.Sprintf(
			"[Error] Caller(%s) call Tripod(%s) Writing(%s) in Block(%s) on Height(%d): %s",
			e.Caller.String(),
			e.TripodName,
			e.WritingName,
			e.BlockHash.String(),
			e.Height,
			e.Err,
		)
	} else {
		str = fmt.Sprintf(
			"[Error] %s Block(%s) on Height(%d) in Tripod(%s): %s",
			e.BlockStage,
			e.BlockHash.String(),
			e.Height,
			e.TripodName,
			e.Err,
		)
	}
	return
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
