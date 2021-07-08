package context

import (
	"bytes"
	"encoding/json"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/yerror"
	"github.com/sirupsen/logrus"
	"strconv"
)

func (c *Context) BindJson(v interface{}) error {
	d := json.NewDecoder(bytes.NewReader([]byte(c.paramsStr)))
	d.UseNumber()
	return d.Decode(v)
}

func (c *Context) Get(name string) interface{} {
	return c.paramsMap[name]
}

func (c *Context) GetHash(name string) Hash {
	h, err := c.TryGetHash(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return h
}

func (c *Context) TryGetHash(name string) (Hash, error) {
	str, err := c.TryGetString(name)
	if err != nil {
		return NullHash, err
	}
	return HexToHash(str), nil
}

func (c *Context) GetAddress(name string) Address {
	a, err := c.TryGetAddress(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return a
}

func (c *Context) TryGetAddress(name string) (Address, error) {
	str, err := c.TryGetString(name)
	if err != nil {
		return NullAddress, err
	}
	return HexToAddress(str), nil
}

func (c *Context) GetString(name string) string {
	str, err := c.TryGetString(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return str
}

func (c *Context) TryGetString(name string) (string, error) {
	pi := c.paramsMap[name]
	if pis, ok := pi.(string); ok {
		return pis, nil
	}
	return "", TypeErr
}

func (c *Context) GetBytes(name string) []byte {
	byt, err := c.TryGetBytes(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return byt
}

func (c *Context) TryGetBytes(name string) ([]byte, error) {
	str, err := c.TryGetString(name)
	return []byte(str), err
}

func (c *Context) GetBoolean(name string) bool {
	b, err := c.TryGetBoolean(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return b
}

func (c *Context) TryGetBoolean(name string) (bool, error) {
	pi := c.paramsMap[name]
	if pis, ok := pi.(bool); ok {
		return pis, nil
	}
	return false, TypeErr
}

func (c *Context) GetInt(name string) int {
	i, err := c.TryGetInt(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *Context) TryGetInt(name string) (int, error) {
	pi := c.getNumberStr(name)
	return strconv.Atoi(pi)
}

func (c *Context) GetUint(name string) uint {
	u, err := c.TryGetUint(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *Context) TryGetUint(name string) (uint, error) {
	u, err := c.TryGetUint64(name)
	return uint(u), err
}

func (c *Context) GetInt8(name string) int8 {
	i, err := c.TryGetInt8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *Context) TryGetInt8(name string) (int8, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 8)
	return int8(i), err
}

func (c *Context) GetUint8(name string) uint8 {
	u, err := c.TryGetUint8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *Context) TryGetUint8(name string) (uint8, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 8)
	return uint8(u), err
}

func (c *Context) GetInt16(name string) int16 {
	i, err := c.TryGetInt16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *Context) TryGetInt16(name string) (int16, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 16)
	return int16(i), err
}

func (c *Context) GetUint16(name string) uint16 {
	u, err := c.TryGetUint16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *Context) TryGetUint16(name string) (uint16, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 16)
	return uint16(u), err
}

func (c *Context) GetInt32(name string) int32 {
	i, err := c.TryGetInt32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *Context) TryGetInt32(name string) (int32, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 32)
	return int32(i), err
}

func (c *Context) GetUint32(name string) uint32 {
	u, err := c.TryGetUint32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *Context) TryGetUint32(name string) (uint32, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 32)
	return uint32(u), err
}

func (c *Context) GetInt64(name string) int64 {
	i, err := c.TryGetInt64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *Context) TryGetInt64(name string) (int64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseInt(pi, 10, 64)
}

func (c *Context) GetUint64(name string) uint64 {
	u, err := c.TryGetUint64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *Context) TryGetUint64(name string) (uint64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseUint(pi, 10, 64)
}

func (c *Context) GetFloat32(name string) float32 {
	f, err := c.TryGetFloat32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (c *Context) TryGetFloat32(name string) (float32, error) {
	pi := c.getNumberStr(name)
	f, err := strconv.ParseFloat(pi, 32)
	return float32(f), err
}

func (c *Context) GetFloat64(name string) float64 {
	f, err := c.TryGetFloat64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (c *Context) TryGetFloat64(name string) (float64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseFloat(pi, 64)
}

func (c *Context) getNumberStr(name string) string {
	return c.paramsMap[name].(json.Number).String()
}
