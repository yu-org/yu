package common

import "yu/context"

type (
	BlockNum uint64
	Execution func(ctx *context.Context) error
	Query func(ctx *context.Context, blockNum BlockNum) error

	Call struct {
		FuncName string
		Params []interface{}
	}
)
