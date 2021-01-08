package context

import (
	"github.com/pkg/errors"
)

type Context struct {
	params map[string]interface{}
}

var TypeError = errors.New("the type of params error")

func NewContext() *Context {
	return &Context{
		params: make(map[string]interface{}),
	}
}
