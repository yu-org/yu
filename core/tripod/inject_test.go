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
	tri := NewTripodWithName("test")
	return &TestInjectTripod{Tripod: tri}
}

type Dayu struct {
	*Tripod
}

func newDayu() *Dayu {
	tri := NewTripodWithName(dayuStr)
	return &Dayu{Tripod: tri}
}

func (*Dayu) String() string {
	return dayuStr
}

type Boyi struct {
	*Tripod
}

func newBoyi() *Boyi {
	tri := NewTripodWithName(boyiStr)
	return &Boyi{Tripod: tri}
}

func (*Boyi) String() string {
	return boyiStr
}

func TestInject(t *testing.T) {
	land := NewLand()

	testTri := newTestTripod()
	testTri.SetLand(land)
	testTri.SetInstance(testTri)

	dayu := newDayu()
	dayu.SetLand(land)
	dayu.SetInstance(dayu)

	boyi := newBoyi()
	boyi.SetLand(land)
	boyi.SetInstance(boyi)

	land.SetTripods(testTri.Tripod, dayu.Tripod, boyi.Tripod)

	err := Inject(testTri)
	assert.NoError(t, err)
	assert.Equal(t, dayuStr, testTri.Dayu.String())
	assert.Equal(t, boyiStr, testTri.boyi.String())
}
