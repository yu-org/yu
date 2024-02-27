package receipt

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
	// codec event
	evResult := NewWithEvents([]*Event{ev})
	byt, err := evResult.Encode()
	assert.NoError(t, err)

	deEvResult := new(Receipt)
	err = deEvResult.Decode(byt)
	assert.NoError(t, err)
	assert.Equal(t, ev, deEvResult.Events[0])
}
