package context

import (
	"github.com/yu-org/yu/common"
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

func (rc *ReadContext) GetBlockHash() *common.Hash {
	return rc.BlockHash
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
