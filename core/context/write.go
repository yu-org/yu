package context

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
)

type WriteContext struct {
	*ParamsResponse

	Block    *types.Block
	Txn      *types.SignedTxn
	TxnIndex int

	Events []*types.Event
	Extra  []byte

	ExtraInterface any

	LeiCost uint64
}

func NewWriteContext(stxn *types.SignedTxn, block *types.Block, idx int) (*WriteContext, error) {
	paramsStr := stxn.Raw.WrCall.Params
	rctx, err := NewParamsResponseFromStr(paramsStr)
	if err != nil {
		return nil, err
	}
	return &WriteContext{
		Txn:            stxn,
		TxnIndex:       idx,
		Block:          block,
		ParamsResponse: rctx,
		Events:         make([]*types.Event, 0),
	}, nil
}

func (c *WriteContext) GetRequestBytes() []byte {
	return []byte(c.Txn.GetParams())
}

func (c *WriteContext) BindJson(v any) error {
	return common.BindJsonParams(c.ParamsStr, v)
}

func (c *WriteContext) GetTimestamp() uint64 {
	return c.Block.Timestamp
}

func (c *WriteContext) GetTxnHash() common.Hash {
	return c.Txn.TxnHash
}

func (c *WriteContext) GetCaller() *common.Address {
	return c.Txn.GetCaller()
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
	event := &types.Event{Value: bytes}
	c.Events = append(c.Events, event)
}

func (c *WriteContext) EmitStringEvent(format string, values ...any) {
	event := &types.Event{Value: []byte(fmt.Sprintf(format, values...))}
	c.Events = append(c.Events, event)
}

func (c *WriteContext) EmitJsonEvent(value any) error {
	byt, err := json.Marshal(value)
	if err != nil {
		logrus.Error("json encode to bytes error: ", err)
		return err
	}
	event := &types.Event{Value: byt}
	c.Events = append(c.Events, event)
	return nil
}

func (c *WriteContext) EmitExtra(extra []byte) {
	c.Extra = extra
}
