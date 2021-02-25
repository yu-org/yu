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

func (l *Land) ExistExec(tripodName, execName string) error {
	t, ok := l.Tripods[tripodName]
	if !ok {
		return TripodNotFound(tripodName)
	}
	ok = t.TripodMeta().ExistExec(execName)
	if !ok {
		return ExecNotFound(execName)
	}
	return nil
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
	err := ctx.SetParams(c.Params)
	if err != nil {
		return err
	}
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
	err := ctx.SetParams(c.Params)
	if err != nil {
		return err
	}
	return qry(ctx, c.BlockNumber)
}
