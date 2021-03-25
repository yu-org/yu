package context

import (
	. "yu/common"
	. "yu/result"
)

type Context struct {
	paramsMap map[string]interface{}
	paramsStr JsonString
	Events    []*Event
	Errors    []*Error
}

func NewContext() *Context {
	return &Context{
		paramsMap: make(map[string]interface{}),
		paramsStr: "",
		Events:    make([]*Event, 0),
		Errors:    make([]*Error, 0),
	}
}

func (c *Context) EmitEvent(value Display) {
	event := &Event{
		Value: value,
	}
	c.Events = append(c.Events, event)
}

func (c *Context) EmitError(e error) {
	err := &Error{
		Err: e,
	}
	c.Errors = append(c.Errors, err)
}
