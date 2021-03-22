package context

import (
	. "yu/common"
	. "yu/result"
)

type Context struct {
	paramsMap map[string]interface{}
	paramsStr JsonString
	Events    []IEvent
	Errors    []IError
}

func NewContext() *Context {
	return &Context{
		paramsMap: make(map[string]interface{}),
		paramsStr: "",
		Events:    make([]IEvent, 0),
		Errors:    make([]IError, 0),
	}
}

func (c *Context) EmitEvent(event IEvent) {
	c.Events = append(c.Events, event)
}

func (c *Context) EmitError(err IError) {
	c.Errors = append(c.Errors, err)
}
