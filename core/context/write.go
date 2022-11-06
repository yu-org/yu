package context

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/codec"
)

type WriteContext struct {
	*ReadContext

	Block *Block
	Txn   *SignedTxn

	Events  []*Event
	Error   *Error
	LeiCost uint64
}

func NewWriteContext(stxn *SignedTxn, block *Block) (*WriteContext, error) {
	paramsStr := stxn.Raw.WrCall.Params
	rctx, err := NewReadContext(paramsStr)
	if err != nil {
		return nil, err
	}
	return &WriteContext{
		Txn:         stxn,
		Block:       block,
		ReadContext: rctx,
		Events:      make([]*Event, 0),
	}, nil
}

func (c *WriteContext) GetCaller() Address {
	return c.Txn.Pubkey.Address()
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

func (c *WriteContext) EmitEvent(value any) error {
	byt, err := codec.GlobalCodec.EncodeToBytes(value)
	if err != nil {
		logrus.Error("encode event to bytes error: ", err)
		return err
	}
	event := &Event{Value: string(byt)}
	c.Events = append(c.Events, event)
	return nil
}

func (c *WriteContext) EmitStringEvent(format string, values ...any) {
	event := &Event{Value: fmt.Sprintf(format, values...)}
	c.Events = append(c.Events, event)
}

func (c *WriteContext) EmitJsonEvent(value any) error {
	byt, err := json.Marshal(value)
	if err != nil {
		logrus.Error("json encode to bytes error: ", err)
		return err
	}
	event := &Event{Value: string(byt)}
	c.Events = append(c.Events, event)
	return nil
}

func (c *WriteContext) EmitError(e error) {
	c.Error = &Error{
		Err: e.Error(),
	}
}
