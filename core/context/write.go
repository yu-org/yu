package context

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/types"
)

type WriteContext struct {
	Block         *Block
	IsConcurrency bool
	// execute txn one by one
	*ParamsResponse
	Txn *SignedTxn

	// execute txn in concurrency
	ParamsResponseList []*ParamsResponse
	TxnList            []*SignedTxn
	ErrorList          []error

	Events []*Event
	Extra  []byte

	LeiCost uint64
}

func NewWriteContextByTxns(stxnList []*SignedTxn, block *Block) (*WriteContext, error) {
	if len(stxnList) < 2 {
		return NewWriteContext(stxnList[0], block)
	}
	c := &WriteContext{
		Block:              block,
		IsConcurrency:      true,
		Events:             make([]*Event, 0),
		ParamsResponseList: make([]*ParamsResponse, 0),
		TxnList:            make([]*SignedTxn, 0),
	}
	for _, stxn := range stxnList {
		paramsStr := stxn.Raw.WrCall.Params
		rctx, err := NewParamsResponseFromStr(paramsStr)
		if err != nil {
			return nil, err
		}
		c.TxnList = append(c.TxnList, stxn)
		c.ParamsResponseList = append(c.ParamsResponseList, rctx)
	}
	c.ErrorList = make([]error, len(c.TxnList), len(c.TxnList))
	return c, nil
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

func (c *WriteContext) GetTxnHash() Hash {
	return c.Txn.TxnHash
}

func (c *WriteContext) GetCaller() *Address {
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

func (c *WriteContext) EmitExtra(extra []byte) {
	c.Extra = extra
}
