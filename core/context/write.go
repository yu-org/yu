package context

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
)

type WriteContext struct {
	*ParamsResponse

	Block *Block
	Txn   *SignedTxn

	Events  []*Event
	Error   *Error
	LeiCost uint64
}

func NewWriteContext(stxn *SignedTxn, block *Block) (*WriteContext, error) {
	paramsStr := stxn.Raw.WrCall.Params
	rctx, err := NewParamsResponseFromStr(paramsStr)
	if err != nil {
		return nil, err
	}
	return &WriteContext{
		Txn:            stxn,
		Block:          block,
		ParamsResponse: rctx,
		Events:         make([]*Event, 0),
	}, nil
}

func (c *WriteContext) BindJson(v any) error {
	return BindJsonParams(c.ParamsStr, v)
}

func (c *WriteContext) GetTimestamp() uint64 {
	return c.Block.Timestamp
}

func (c *WriteContext) GetCaller() Address {
	return c.Txn.GetCallerAddr()
}

func (c *WriteContext) FromP2P() bool {
	return c.Txn.FromP2p()
}

func (c *WriteContext) SetLei(lei uint64) {
	c.LeiCost = lei
}

func (c *WriteContext) SetLeiFn(fn func() uint64) {
	c.LeiCost = fn()
}

func (c *WriteContext) EmitEvent(bytes []byte) {
	event := &Event{Value: bytes}
	c.Events = append(c.Events, event)
}

func (c *WriteContext) EmitStringEvent(format string, values ...any) {
	event := &Event{Value: []byte(fmt.Sprintf(format, values...))}
	c.Events = append(c.Events, event)
}

func (c *WriteContext) EmitJsonEvent(value any) error {
	byt, err := json.Marshal(value)
	if err != nil {
		logrus.Error("json encode to bytes error: ", err)
		return err
	}
	event := &Event{Value: byt}
	c.Events = append(c.Events, event)
	return nil
}

func (c *WriteContext) EmitError(e error) {
	c.Error = &Error{
		Err: e.Error(),
	}
}
