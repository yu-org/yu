package context

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	"strconv"
)

type ReadContext struct {
	paramsStr string
	paramsMap map[string]interface{}

	response []byte
}

func NewReadContext(paramsStr string) (*ReadContext, error) {
	var paramsMap = make(map[string]interface{})
	err := BindJsonParams(paramsStr, &paramsMap)
	if err != nil {
		return nil, err
	}
	return &ReadContext{
		paramsStr: paramsStr,
		paramsMap: paramsMap,
	}, nil
}

func (c *ReadContext) Bindjson(v any) error {
	return BindJsonParams(c.paramsStr, v)
}

func (c *ReadContext) Response() []byte {
	return c.response
}

func (c *ReadContext) Bytes(byt []byte) {
	c.response = byt
}

func (c *ReadContext) String(format string, values ...any) {
	c.response = []byte(fmt.Sprintf(format, values...))
}

type H map[string]interface{}

func (c *ReadContext) Json(v any) (err error) {
	c.response, err = json.Marshal(v)
	return
}

func (c *ReadContext) Get(name string) interface{} {
	return c.paramsMap[name]
}

func (c *ReadContext) GetHash(name string) Hash {
	h, err := c.TryGetHash(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return h
}

func (c *ReadContext) TryGetHash(name string) (Hash, error) {
	str, err := c.TryGetString(name)
	if err != nil {
		return NullHash, err
	}
	return HexToHash(str), nil
}

func (c *ReadContext) GetAddress(name string) Address {
	a, err := c.TryGetAddress(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return a
}

func (c *ReadContext) TryGetAddress(name string) (Address, error) {
	str, err := c.TryGetString(name)
	if err != nil {
		return NullAddress, err
	}
	return HexToAddress(str), nil
}

func (c *ReadContext) GetString(name string) string {
	str, err := c.TryGetString(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return str
}

func (c *ReadContext) TryGetString(name string) (string, error) {
	pi := c.paramsMap[name]
	if pis, ok := pi.(string); ok {
		return pis, nil
	}
	return "", TypeErr
}

func (c *ReadContext) GetBytes(name string) []byte {
	byt, err := c.TryGetBytes(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return byt
}

func (c *ReadContext) TryGetBytes(name string) ([]byte, error) {
	str, err := c.TryGetString(name)
	return []byte(str), err
}

func (c *ReadContext) GetBoolean(name string) bool {
	b, err := c.TryGetBoolean(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return b
}

func (c *ReadContext) TryGetBoolean(name string) (bool, error) {
	pi := c.paramsMap[name]
	if pis, ok := pi.(bool); ok {
		return pis, nil
	}
	return false, TypeErr
}

func (c *ReadContext) GetInt(name string) int {
	i, err := c.TryGetInt(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *ReadContext) TryGetInt(name string) (int, error) {
	pi := c.getNumberStr(name)
	return strconv.Atoi(pi)
}

func (c *ReadContext) GetUint(name string) uint {
	u, err := c.TryGetUint(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *ReadContext) TryGetUint(name string) (uint, error) {
	u, err := c.TryGetUint64(name)
	return uint(u), err
}

func (c *ReadContext) GetInt8(name string) int8 {
	i, err := c.TryGetInt8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *ReadContext) TryGetInt8(name string) (int8, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 8)
	return int8(i), err
}

func (c *ReadContext) GetUint8(name string) uint8 {
	u, err := c.TryGetUint8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *ReadContext) TryGetUint8(name string) (uint8, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 8)
	return uint8(u), err
}

func (c *ReadContext) GetInt16(name string) int16 {
	i, err := c.TryGetInt16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *ReadContext) TryGetInt16(name string) (int16, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 16)
	return int16(i), err
}

func (c *ReadContext) GetUint16(name string) uint16 {
	u, err := c.TryGetUint16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *ReadContext) TryGetUint16(name string) (uint16, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 16)
	return uint16(u), err
}

func (c *ReadContext) GetInt32(name string) int32 {
	i, err := c.TryGetInt32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *ReadContext) TryGetInt32(name string) (int32, error) {
	pi := c.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 32)
	return int32(i), err
}

func (c *ReadContext) GetUint32(name string) uint32 {
	u, err := c.TryGetUint32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *ReadContext) TryGetUint32(name string) (uint32, error) {
	pi := c.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 32)
	return uint32(u), err
}

func (c *ReadContext) GetInt64(name string) int64 {
	i, err := c.TryGetInt64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (c *ReadContext) TryGetInt64(name string) (int64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseInt(pi, 10, 64)
}

func (c *ReadContext) GetUint64(name string) uint64 {
	u, err := c.TryGetUint64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (c *ReadContext) TryGetUint64(name string) (uint64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseUint(pi, 10, 64)
}

func (c *ReadContext) GetFloat32(name string) float32 {
	f, err := c.TryGetFloat32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (c *ReadContext) TryGetFloat32(name string) (float32, error) {
	pi := c.getNumberStr(name)
	f, err := strconv.ParseFloat(pi, 32)
	return float32(f), err
}

func (c *ReadContext) GetFloat64(name string) float64 {
	f, err := c.TryGetFloat64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (c *ReadContext) TryGetFloat64(name string) (float64, error) {
	pi := c.getNumberStr(name)
	return strconv.ParseFloat(pi, 64)
}

func (c *ReadContext) getNumberStr(name string) string {
	return c.paramsMap[name].(json.Number).String()
}
