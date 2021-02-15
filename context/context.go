package context

import . "yu/common"

type Context struct {
	paramsMap map[string]interface{}
	paramsStr JsonString
}

func NewContext() *Context {
	return &Context{
		paramsMap: make(map[string]interface{}),
		paramsStr: "",
	}
}
