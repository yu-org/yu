package poa

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"testing"
)

func TestSig(t *testing.T) {
	wrCall := common.WrCall{
		TripodName: "question",
		FuncName:   "AddQuestion",
		Params:     `{"title":"aaa","content":"bbb","timestamp":1692652052812}`,
		LeiPrice:   0,
		Tips:       0,
	}
	t.Log(wrCall)
	hash, err := wrCall.Hash()
	assert.NoError(t, err)
	t.Logf("%x", hash)
}
