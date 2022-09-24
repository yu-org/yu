package context

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/result"
	. "github.com/yu-org/yu/core/types"
	"github.com/yu-org/yu/utils/codec"
)

type Context struct {
	Caller    Address
	Block     *Block
	paramsMap map[string]interface{}
	paramsStr string
	Events    []*Event
	Error     *Error
	LeiCost   uint64
}

func NewContext(caller Address, paramsStr string, block *Block) (*Context, error) {
	var i interface{}
	d := json.NewDecoder(bytes.NewReader([]byte(paramsStr)))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return nil, err
	}
	return &Context{
		Caller:    caller,
		Block:     block,
		paramsMap: i.(map[string]interface{}),
		paramsStr: paramsStr,
		Events:    make([]*Event, 0),
	}, nil
}

func (c *Context) SetLei(lei uint64) {
	c.LeiCost = lei
}

func (c *Context) SetLeiFn(fn func() uint64) {
	c.LeiCost = fn()
}

func (c *Context) EmitEvent(value interface{}) error {
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

func (c *Context) EmitError(e error) {
	c.Error = &Error{
		Err: e.Error(),
	}
}
