package context

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/common/yerror"
	"net/http"
)

type ReadContext struct {
	BlockHash *common.Hash
	rdCall    *common.RdCall
	resp      *ResponseData
}

type ResponseData struct {
	StatusCode int

	DataInterface any
	IsJson        bool

	ContentType string
	DataBytes   []byte
}

func NewReadContext(rdCall *common.RdCall) (*ReadContext, error) {
	var blockHash *common.Hash
	if rdCall.BlockHash != "" {
		blockH := common.HexToHash(rdCall.BlockHash)
		blockHash = &blockH
	}

	return &ReadContext{
		BlockHash: blockHash,
		rdCall:    rdCall,
	}, nil
}

func (rc *ReadContext) Response() *ResponseData {
	return rc.resp
}

func (rc *ReadContext) BindJson(v any) error {
	return common.BindJsonParams(rc.rdCall.Params, v)
}

func (rc *ReadContext) GetParams(key string) any {
	value, err := rc.TryGetParams(key)
	if err != nil {
		logrus.Panicf("ReadContext.GetParams() failed: %v", err)
	}
	return value
}

func (rc *ReadContext) TryGetParams(key string) (any, error) {
	params := make(map[string]any)
	err := rc.BindJson(params)
	if err != nil {
		return nil, err
	}
	return params[key], nil
}

func (rc *ReadContext) GetString(key string) string {
	value, err := rc.TryGetString(key)
	if err != nil {
		logrus.Panicf("ReadContext.GetString() failed: %v", err)
	}
	return value
}

func (rc *ReadContext) TryGetString(key string) (string, error) {
	params := make(map[string]any)
	err := rc.BindJson(params)
	if err != nil {
		return "", err
	}
	value := params[key]
	if _, ok := value.(string); !ok {
		return "", yerror.TypeErr
	}
	return value.(string), nil
}

func (rc *ReadContext) GetBlockHash() *common.Hash {
	return rc.BlockHash
}

func (rc *ReadContext) GetParams() string {
	return rc.rdCall.Params
}

func (rc *ReadContext) Json(code int, v any) {
	rc.resp = &ResponseData{
		StatusCode:    code,
		DataInterface: v,
		IsJson:        true,
	}
}

func (rc *ReadContext) JsonOk(v any) {
	rc.Json(http.StatusOK, v)
	// rc.JSON(http.StatusOK, v)
}

func (rc *ReadContext) Data(code int, contentType string, data []byte) {
	rc.resp = &ResponseData{
		StatusCode:  code,
		ContentType: contentType,
		DataBytes:   data,
	}
	// rc.Data(http.StatusOK, contentType, data)
}

func (rc *ReadContext) DataOk(contentType string, data []byte) {
	rc.Data(http.StatusOK, contentType, data)
}

func (rc *ReadContext) Err(code int, err error) {
	rc.Json(code, struct {
		Err error `json:"err"`
	}{Err: err})
}

func (rc *ReadContext) ErrOk(err error) {
	rc.Err(http.StatusOK, err)
}
