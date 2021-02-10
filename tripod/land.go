package tripod

import (
	. "yu/common"
	. "yu/context"
	. "yu/yerror"
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
		return TripodNotFound(c.TripodName)
	}
	ph := Tripod.TripodMeta()
	fn := ph.GetExec(c.ExecName)
	if fn == nil {
		return ExecNotFound(c.ExecName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params.Params)
	return fn(ctx)
}

func (l *Land) Query(c *Qcall) error {
	Tripod, ok := l.Tripods[c.TripodName]
	if !ok {
		return TripodNotFound(c.TripodName)
	}
	ph := Tripod.TripodMeta()
	qry := ph.GetQuery(c.QueryName)
	if qry == nil {
		return QryNotFound(c.QueryName)
	}
	ctx := NewContext()
	ctx.SetParams(c.Params.Params)
	return qry(ctx, c.Params.BlockNumber)
}
