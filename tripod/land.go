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

func (l *Land) SetTripods(Tripods ...Tripod) {
	for _, Tripod := range Tripods {
		TripodName := Tripod.TripodMeta().Name()
		l.Tripods[TripodName] = Tripod
	}
}

func (l *Land) Execute(c *Ecall) error {
	Tripod, ok := l.Tripods[c.TripodName]
	if !ok {
		return errors.Errorf("Tripod (%s) not found", c.TripodName)
	}
	ph := Tripod.TripodMeta()
	fn := ph.GetExec(c.ExecName)
	if fn == nil {
		return errors.Errorf("Execution (%s) not found", c.ExecName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params.Params)
	return fn(ctx)
}

func (l *Land) Query(c *Qcall) error {
	Tripod, ok := l.Tripods[c.TripodName]
	if !ok {
		return errors.Errorf("Tripod (%s) not found", c.TripodName)
	}
	ph := Tripod.TripodMeta()
	qry := ph.GetQuery(c.QueryName)
	if qry == nil {
		return errors.Errorf("Query (%s) not found", c.QueryName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params.Params)
	return qry(ctx, c.Params.BlockNumber)
}
