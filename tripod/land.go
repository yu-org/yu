package tripod

import (
	. "yu/common"
	. "yu/context"
	. "yu/yerror"
)

type Land struct {
	orderedTripods []Tripod
	// Key: the Name of Tripod
	tripodsMap map[string]Tripod
}

func NewLand() *Land {
	return &Land{
		tripodsMap:     make(map[string]Tripod),
		orderedTripods: make([]Tripod, 0),
	}
}

func (l *Land) SetTripods(Tripods ...Tripod) {
	for _, Tripod := range Tripods {
		TripodName := Tripod.TripodMeta().Name()
		l.tripodsMap[TripodName] = Tripod

		l.orderedTripods = append(l.orderedTripods, Tripod)
	}
}

func (l *Land) ExistExec(tripodName, execName string) error {
	t, ok := l.tripodsMap[tripodName]
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
	Tripod, ok := l.tripodsMap[c.TripodName]
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
	Tripod, ok := l.tripodsMap[c.TripodName]
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

func (l *Land) RangeMap(fn func(string, Tripod) error) error {
	for name, tri := range l.tripodsMap {
		err := fn(name, tri)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Land) RangeList(fn func(Tripod) error) error {
	for _, tri := range l.orderedTripods {
		err := fn(tri)
		if err != nil {
			return err
		}
	}
	return nil
}
