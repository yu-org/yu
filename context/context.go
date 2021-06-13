package context

import (
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/result"
	"github.com/Lawliet-Chan/yu/utils/codec"
	"github.com/sirupsen/logrus"
)

type Context struct {
	Caller    Address
	paramsMap map[string]interface{}
	paramsStr JsonString
	Events    []*Event
	Error     *Error
}

func NewContext(caller Address, paramsStr JsonString) (*Context, error) {
	pMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(paramsStr), &pMap)
	if err != nil {
		return nil, err
	}
	return &Context{
		Caller:    caller,
		paramsMap: pMap,
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
		Err: e,
	}
}
