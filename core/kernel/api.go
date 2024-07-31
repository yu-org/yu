package kernel

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/types"
	"net/http"
	"strconv"
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
	receipts, err := k.getReceipts(ctx)
	if err != nil {
		RenderError(ctx, ReceiptFailure, err)
		return
	}
	RenderSuccess(ctx, receipts)
}

func (k *Kernel) GetReceiptsCount(ctx *gin.Context) {
	receipts, err := k.getReceipts(ctx)
	if err != nil {
		RenderError(ctx, ReceiptFailure, err)
		return
	}
	RenderSuccess(ctx, len(receipts))
}

func (k *Kernel) getReceipts(ctx *gin.Context) ([]*types.Receipt, error) {
	var (
		block *types.CompactBlock
		err   error
	)

	blockHashStr := ctx.Query("block_hash")
	blockNumberStr := ctx.Query("block_number")
	if blockNumberStr != "" {
		blockNumber, err := strconv.Atoi(blockNumberStr)
		if err != nil {
			return nil, err
		}
		block, err = k.Chain.GetCompactBlockByHeight(common.BlockNum(blockNumber))
		if err != nil {
			return nil, err
		}
	} else if blockHashStr != "" {
		blockHash := common.HexToHash(blockHashStr)
		block, err = k.Chain.GetCompactBlock(blockHash)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("none request params")
	}

	var receipts []*types.Receipt
	for _, txHash := range block.TxnsHashes {
		receipt, err := k.TxDB.GetReceipt(txHash)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}
