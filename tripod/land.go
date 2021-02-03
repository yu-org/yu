package tripod

import (
	"github.com/pkg/errors"
	. "yu/common"
	. "yu/context"
)

type Land struct {
	// Key: the Name of Tripod
	Tripods map[string]Tripod
}

func NewLand() *Land {
	return &Land{
		Tripods: make(map[string]Tripod),
	}
}

func (s *Land) SetTripods(Tripods ...Tripod) {
	for _, Tripod := range Tripods {
		TripodName := Tripod.TripodMeta().Name()
		s.Tripods[TripodName] = Tripod
	}
}

func (s *Land) Execute(c *Call) error {
	Tripod, ok := s.Tripods[c.TripodName]
	if !ok {
		return errors.Errorf("Tripod (%s) not found", c.TripodName)
	}
	ph := Tripod.TripodMeta()
	fn := ph.GetExecFn(c.FuncName)
	if fn == nil {
		return errors.Errorf("Execution (%s) not found", c.FuncName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params)
	return fn(ctx)
}
