package context

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/result"
	"github.com/yu-org/yu/utils/codec"
)

type Context struct {
	Caller    Address
	paramsMap map[string]interface{}
	paramsStr JsonString
	Events    []*Event
	Error     *Error
}

func NewContext(caller Address, paramsStr JsonString) (*Context, error) {
	var i interface{}
	d := json.NewDecoder(bytes.NewReader([]byte(paramsStr)))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return nil, err
	}
	return &Context{
		Caller:    caller,
		paramsMap: i.(map[string]interface{}),
		paramsStr: paramsStr,
		Events:    make([]*Event, 0),
	}, nil
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
