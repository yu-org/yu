package context

import (
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
	"reflect"
)

func (c *Context) BindJson(v interface{}) error {
	return json.Unmarshal([]byte(c.paramsStr), v)
}

func (c *Context) Get(name string) interface{} {
	return c.paramsMap[name]
}

func (c *Context) GetString(name string) string {
	str, err := c.TryGetString(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return str
}

func (c *Context) TryGetString(name string) (string, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.String {
		return pi.(string), nil
	}
	return "", TypeErr
}

func (c *Context) GetBytes(name string) []byte {
	byt, err := c.TryGetBytes(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return byt
}

func (c *Context) TryGetBytes(name string) ([]byte, error) {
	pi := c.paramsMap[name]
	if byt, ok := pi.([]byte); ok {
		return byt, nil
	}
	return nil, TypeErr
}

func (c *Context) GetBoolean(name string) bool {
	b, err := c.TryGetBoolean(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return b
}

func (c *Context) TryGetBoolean(name string) (bool, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Bool {
		return pi.(bool), nil
	}
	return false, TypeErr
}

func (c *Context) GetInt(name string) int {
	i, err := c.TryGetInt(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, TypeErr.Error())
	}
	return i
}

func (c *Context) TryGetInt(name string) (int, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int {
		return pi.(int), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint(name string) uint {
	u, err := c.TryGetUint(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return u
}

func (c *Context) TryGetUint(name string) (uint, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint {
		return pi.(uint), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt8(name string) int8 {
	i, err := c.TryGetInt8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return i
}

func (c *Context) TryGetInt8(name string) (int8, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int8 {
		return pi.(int8), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint8(name string) uint8 {
	u, err := c.TryGetUint8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return u
}

func (c *Context) TryGetUint8(name string) (uint8, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint8 {
		return pi.(uint8), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt16(name string) int16 {
	i, err := c.TryGetInt16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return i
}

func (c *Context) TryGetInt16(name string) (int16, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int16 {
		return pi.(int16), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint16(name string) uint16 {
	u, err := c.TryGetUint16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return u
}

func (c *Context) TryGetUint16(name string) (uint16, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint16 {
		return pi.(uint16), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt32(name string) int32 {
	i, err := c.TryGetInt32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return i
}

func (c *Context) TryGetInt32(name string) (int32, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int32 {
		return pi.(int32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint32(name string) uint32 {
	u, err := c.TryGetUint32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return u
}

func (c *Context) TryGetUint32(name string) (uint32, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint32 {
		return pi.(uint32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetInt64(name string) int64 {
	i, err := c.TryGetInt64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return i
}

func (c *Context) TryGetInt64(name string) (int64, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Int64 {
		return pi.(int64), nil
	}
	return 0, TypeErr
}

func (c *Context) GetUint64(name string) uint64 {
	u, err := c.TryGetUint64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return u
}

func (c *Context) TryGetUint64(name string) (uint64, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Uint64 {
		return pi.(uint64), nil
	}
	return 0, TypeErr
}

func (c *Context) GetFloat32(name string) float32 {
	f, err := c.TryGetFloat32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return f
}

func (c *Context) TryGetFloat32(name string) (float32, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Float32 {
		return pi.(float32), nil
	}
	return 0, TypeErr
}

func (c *Context) GetFloat64(name string) float64 {
	f, err := c.TryGetFloat64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s",name, TypeErr.Error())
	}
	return f
}

func (c *Context) TryGetFloat64(name string) (float64, error) {
	pi := c.paramsMap[name]
	if reflect.TypeOf(pi).Kind() == reflect.Float64 {
		return pi.(float64), nil
	}
	return 0, TypeErr
}
