package tripod

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	dayuStr = "dayu"
	boyiStr = "boyi"
)

type TestInjectTripod struct {
	*Tripod
	Dayu *Dayu `tripod:"dayu"`
	boyi *Boyi `tripod:"boyi"`
}

func newTestTripod() *TestInjectTripod {
	tri := NewTripod("test")
	return &TestInjectTripod{Tripod: tri}
}

type Dayu struct {
	*Tripod
}

func newDayu() *Dayu {
	tri := NewTripod(dayuStr)
	return &Dayu{Tripod: tri}
}

func (*Dayu) String() string {
	return dayuStr
}

type Boyi struct {
	*Tripod
}

func newBoyi() *Boyi {
	tri := NewTripod(boyiStr)
	return &Boyi{Tripod: tri}
}

func (*Boyi) String() string {
	return boyiStr
}

func TestInject(t *testing.T) {
	land := NewLand()

	tri := newTestTripod()
	tri.SetLand(land)
	tri.SetInstance(tri)

	dayu := newDayu()
	dayu.SetLand(land)
	dayu.SetInstance(dayu)

	boyi := newBoyi()
	boyi.SetLand(land)
	boyi.SetInstance(boyi)

	land.SetTripods(tri.Tripod, dayu.Tripod, boyi.Tripod)

	err := Inject(tri)
	assert.NoError(t, err)
	assert.Equal(t, dayuStr, tri.Dayu.String())
	assert.Equal(t, boyiStr, tri.boyi.String())
}
