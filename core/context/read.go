package context

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"net/http"
)

type ReadContext struct {
	*gin.Context
	BlockHash *common.Hash
	rdCall    *common.RdCall
}

func NewReadContext(ctx *gin.Context) (*ReadContext, error) {
	rdCall := new(common.RdCall)
	err := ctx.BindJSON(rdCall)
	if err != nil {
		return nil, err
	}
	logrus.Info("new--read--context")
	var blockHash *common.Hash
	if rdCall.BlockHash != "" {
		blockH := common.HexToHash(rdCall.BlockHash)
		blockHash = &blockH
	}

	return &ReadContext{
		Context:   ctx,
		BlockHash: blockHash,
		rdCall:    rdCall,
	}, nil
}

func (rc *ReadContext) BindJson(v any) error {
	return common.BindJsonParams(rc.rdCall.Params, v)
}

func (rc *ReadContext) GetBlockHash() *common.Hash {
	return rc.BlockHash
}

func (rc *ReadContext) JsonOk(v any) {
	rc.JSON(http.StatusOK, v)
}

func (rc *ReadContext) StringOk(format string, values ...any) {
	rc.String(http.StatusOK, format, values)
}

func (rc *ReadContext) DataOk(contentType string, data []byte) {
	rc.Data(http.StatusOK, contentType, data)
}

func (rc *ReadContext) ErrOk(err error) {
	rc.StringOk(err.Error())
}
