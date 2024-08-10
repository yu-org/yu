package protocol

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

var (
	Success = 0

	BlockFailure   = 10001
	TxnFailure     = 10002
	ReceiptFailure = 10003
)

type APIResponse struct {
	Code   int    `json:"code"`
	ErrMsg string `json:"err_msg"`
	Data   any    `json:"data"`
}

func (a *APIResponse) IsSuccess() bool {
	return a.Code == Success
}

func (a *APIResponse) Error() error {
	return errors.New(a.ErrMsg)
}

func RenderSuccess(ctx *gin.Context, data any) {
	RenderJson(ctx, Success, nil, data)
}

func RenderError(ctx *gin.Context, code int, err error) {
	RenderJson(ctx, code, err, nil)
}

func RenderJson(ctx *gin.Context, code int, err error, data any) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	resp := APIResponse{
		Code:   code,
		ErrMsg: errMsg,
		Data:   data,
	}
	ctx.JSON(http.StatusOK, resp)
}
