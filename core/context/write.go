package context

import (
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

func (c *WriteContext) EmitEvent(value interface{}) error {
	byt, err := codec.GlobalCodec.EncodeToBytes(value)
	if err != nil {
		logrus.Errorf("encode event to bytes error: %s", err.Error())
		return err
	}
	event := &Event{
		Value: string(byt),
	}
	c.Events = append(c.Events, event)
	return nil
}

func (c *WriteContext) EmitError(e error) {
	c.Error = &Error{
		Err: e.Error(),
	}
}
