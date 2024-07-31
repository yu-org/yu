package kernel

import (
	"github.com/gin-gonic/gin"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
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

func (k *Kernel) GetReceipts(ctx *gin.Context) {
	blockHashStr := ctx.GetString("block_hash")
	blockHash := common.HexToHash(blockHashStr)
	block, err := k.Chain.GetBlock(blockHash)
	if err != nil {
		RenderError(ctx, BlockFailure, err)
		return
	}
	var receipts []*types.Receipt
	for _, txHash := range block.Compact().TxnsHashes {
		receipt, err := k.TxDB.GetReceipt(txHash)
		if err != nil {
			RenderError(ctx, ReceiptFailure, err)
			return
		}
		receipts = append(receipts, receipt)
	}
	RenderSuccess(ctx, receipts)
}
