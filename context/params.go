package context

import (
	"reflect"
	. "yu/yerror"
)

func (c *Context) SetParams(params []interface{}) {
	for _, param := range params {
		// todo: fix it. it should be a name of param.
		typ := reflect.TypeOf(param)
		name := typ.Name()
		c.params[name] = param
	}
}

func (c *Context) Get(name string) interface{} {
	return c.params[name]
}

func (c *Context) GetString(name string) (string, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.String {
		return pi.(string), nil
	}
	return "", TypeErr
}

func (c *Context) GetBoolean(name string) (bool, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Bool {
		return pi.(bool), nil
	}
	return false, TypeErr
}

func (c *Context) GetInt(name string) (int, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int {
		return pi.(int), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint(name string) (uint, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint {
		return pi.(uint), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt8(name string) (int8, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int8 {
		return pi.(int8), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint8(name string) (uint8, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint8 {
		return pi.(uint8), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt16(name string) (int16, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int16 {
		return pi.(int16), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint16(name string) (uint16, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint16 {
		return pi.(uint16), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt32(name string) (int32, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int32 {
		return pi.(int32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint32(name string) (uint32, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint32 {
		return pi.(uint32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt64(name string) (int64, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int64 {
		return pi.(int64), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint64(name string) (uint64, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint64 {
		return pi.(uint64), nil
	}
	return 0, TypeErr
}

func (c *Context) GetFloat32(name string) (float32, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Float32 {
		return pi.(float32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetFloat64(name string) (float64, error) {
	pi := c.params[name]
	if reflect.TypeOf(pi).Kind() == reflect.Float64 {
		return pi.(float64), nil
	}
	return 0, TypeErr
}
