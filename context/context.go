package context

import (
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/result"
)

type Context struct {
	Caller    Address
	paramsMap map[string]interface{}
	paramsStr JsonString
	Events    []*Event
	Errors    []*Error
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
		Errors:    make([]*Error, 0),
	}, nil
}

func (c *Context) EmitEvent(value Display) error {
	str, err := value.ToString()
	if err != nil {
		return err
	}
	event := &Event{
		Value: str,
	}
	c.Events = append(c.Events, event)
	return nil
}

func (c *Context) EmitError(e error) {
	err := &Error{
		Err: e,
	}
	c.Errors = append(c.Errors, err)
}
