package context

type Context struct {
	params map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		params: make(map[string]interface{}),
	}
}
