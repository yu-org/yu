package result

import (
	"github.com/stretchr/testify/assert"
	"github.com/yu-org/yu/common"
	"testing"
)

var (
	ev = &Event{
		Caller:      common.NullAddress,
		BlockStage:  "",
		BlockHash:   common.NullHash,
		Height:      10,
		TripodName:  "event_tripod",
		WritingName: "event_wr",
		Value:       nil,
		LeiCost:     0,
	}
	er = &Error{
		Caller:      common.NullAddress,
		BlockStage:  "",
		BlockHash:   common.NullHash,
		Height:      9,
		TripodName:  "error-tripod",
		WritingName: "error-wr",
		Err:         "",
	}
)

func TestCodecResult(t *testing.T) {
	// codec event
	evResult := NewEvent(ev)
	byt, err := evResult.Encode()
	assert.NoError(t, err)

	deEvResult := new(Result)
	err = deEvResult.Decode(byt)
	assert.NoError(t, err)
	assert.Equal(t, ev, deEvResult.Event)

	// codec error
	erResult := NewError(er)
	byt, err = erResult.Encode()
	assert.NoError(t, err)

	deErResult := new(Result)
	err = deErResult.Decode(byt)
	assert.NoError(t, err)
	assert.Equal(t, er, deErResult.Error)
}
