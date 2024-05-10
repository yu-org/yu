package context

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	"strconv"
)

type ParamsResponse struct {
	ParamsStr string
	paramsMap map[string]interface{}

	response []byte
}

func NewParamsResponseFromStr(paramsStr string) (*ParamsResponse, error) {
	var paramsMap = make(map[string]interface{})
	err := BindJsonParams(paramsStr, &paramsMap)
	if err != nil {
		return nil, err
	}
	return &ParamsResponse{
		ParamsStr: paramsStr,
		paramsMap: paramsMap,
	}, nil
}

func (rc *ParamsResponse) Response() []byte {
	return rc.response
}

func (rc *ParamsResponse) Bytes(byt []byte) {
	rc.response = byt
}

func (rc *ParamsResponse) String(format string, values ...any) {
	rc.response = []byte(fmt.Sprintf(format, values...))
}

type H map[string]interface{}

func (rc *ParamsResponse) Json(v any) (err error) {
	rc.response, err = json.Marshal(v)
	return
}

func (rc *ParamsResponse) Get(name string) interface{} {
	return rc.paramsMap[name]
}

func (rc *ParamsResponse) GetHash(name string) Hash {
	h, err := rc.TryGetHash(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return h
}

func (rc *ParamsResponse) TryGetHash(name string) (Hash, error) {
	str, err := rc.TryGetString(name)
	if err != nil {
		return NullHash, err
	}
	return HexToHash(str), nil
}

func (rc *ParamsResponse) GetAddress(name string) *Address {
	a, err := rc.TryGetAddress(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return &a
}

func (rc *ParamsResponse) TryGetAddress(name string) (Address, error) {
	str, err := rc.TryGetString(name)
	if err != nil {
		return NullAddress, err
	}
	return HexToAddress(str), nil
}

func (rc *ParamsResponse) GetString(name string) string {
	str, err := rc.TryGetString(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return str
}

func (rc *ParamsResponse) TryGetString(name string) (string, error) {
	pi := rc.paramsMap[name]
	if pis, ok := pi.(string); ok {
		return pis, nil
	}
	return "", TypeErr
}

func (rc *ParamsResponse) GetBytes(name string) []byte {
	byt, err := rc.TryGetBytes(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return byt
}

func (rc *ParamsResponse) TryGetBytes(name string) ([]byte, error) {
	str, err := rc.TryGetString(name)
	return []byte(str), err
}

func (rc *ParamsResponse) GetBoolean(name string) bool {
	b, err := rc.TryGetBoolean(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return b
}

func (rc *ParamsResponse) TryGetBoolean(name string) (bool, error) {
	pi := rc.paramsMap[name]
	if pis, ok := pi.(bool); ok {
		return pis, nil
	}
	return false, TypeErr
}

func (rc *ParamsResponse) GetInt(name string) int {
	i, err := rc.TryGetInt(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (rc *ParamsResponse) TryGetInt(name string) (int, error) {
	pi := rc.getNumberStr(name)
	return strconv.Atoi(pi)
}

func (rc *ParamsResponse) GetUint(name string) uint {
	u, err := rc.TryGetUint(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (rc *ParamsResponse) TryGetUint(name string) (uint, error) {
	u, err := rc.TryGetUint64(name)
	return uint(u), err
}

func (rc *ParamsResponse) GetInt8(name string) int8 {
	i, err := rc.TryGetInt8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (rc *ParamsResponse) TryGetInt8(name string) (int8, error) {
	pi := rc.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 8)
	return int8(i), err
}

func (rc *ParamsResponse) GetUint8(name string) uint8 {
	u, err := rc.TryGetUint8(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (rc *ParamsResponse) TryGetUint8(name string) (uint8, error) {
	pi := rc.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 8)
	return uint8(u), err
}

func (rc *ParamsResponse) GetInt16(name string) int16 {
	i, err := rc.TryGetInt16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (rc *ParamsResponse) TryGetInt16(name string) (int16, error) {
	pi := rc.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 16)
	return int16(i), err
}

func (rc *ParamsResponse) GetUint16(name string) uint16 {
	u, err := rc.TryGetUint16(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (rc *ParamsResponse) TryGetUint16(name string) (uint16, error) {
	pi := rc.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 16)
	return uint16(u), err
}

func (rc *ParamsResponse) GetInt32(name string) int32 {
	i, err := rc.TryGetInt32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (rc *ParamsResponse) TryGetInt32(name string) (int32, error) {
	pi := rc.getNumberStr(name)
	i, err := strconv.ParseInt(pi, 10, 32)
	return int32(i), err
}

func (rc *ParamsResponse) GetUint32(name string) uint32 {
	u, err := rc.TryGetUint32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (rc *ParamsResponse) TryGetUint32(name string) (uint32, error) {
	pi := rc.getNumberStr(name)
	u, err := strconv.ParseUint(pi, 10, 32)
	return uint32(u), err
}

func (rc *ParamsResponse) GetInt64(name string) int64 {
	i, err := rc.TryGetInt64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return i
}

func (rc *ParamsResponse) TryGetInt64(name string) (int64, error) {
	pi := rc.getNumberStr(name)
	return strconv.ParseInt(pi, 10, 64)
}

func (rc *ParamsResponse) GetUint64(name string) uint64 {
	u, err := rc.TryGetUint64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return u
}

func (rc *ParamsResponse) TryGetUint64(name string) (uint64, error) {
	pi := rc.getNumberStr(name)
	return strconv.ParseUint(pi, 10, 64)
}

func (rc *ParamsResponse) GetFloat32(name string) float32 {
	f, err := rc.TryGetFloat32(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (rc *ParamsResponse) TryGetFloat32(name string) (float32, error) {
	pi := rc.getNumberStr(name)
	f, err := strconv.ParseFloat(pi, 32)
	return float32(f), err
}

func (rc *ParamsResponse) GetFloat64(name string) float64 {
	f, err := rc.TryGetFloat64(name)
	if err != nil {
		logrus.Panicf("get param(%s) error: %s", name, err.Error())
	}
	return f
}

func (rc *ParamsResponse) TryGetFloat64(name string) (float64, error) {
	pi := rc.getNumberStr(name)
	return strconv.ParseFloat(pi, 64)
}

func (rc *ParamsResponse) getNumberStr(name string) string {
	return rc.paramsMap[name].(json.Number).String()
}
