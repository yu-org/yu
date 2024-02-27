package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	ev = &Event{
		Value: nil,
	}
)

func TestCodecResult(t *testing.T) {
	// codec receipt
	evReceipt := &Receipt{Events: []*Event{ev}}
	byt, err := evReceipt.Encode()
	assert.NoError(t, err)

	deEvReceipt := new(Receipt)
	err = deEvReceipt.Decode(byt)
	assert.NoError(t, err)
	assert.Equal(t, ev, deEvReceipt.Events[0])
}
