package tripod

import (
	. "github.com/Lawliet-Chan/yu/chain_env"
	. "github.com/Lawliet-Chan/yu/common"
	. "github.com/Lawliet-Chan/yu/context"
	. "github.com/Lawliet-Chan/yu/yerror"
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
	for _, tri := range Tripods {
		triName := tri.GetTripodMeta().Name()
		l.tripodsMap[triName] = tri

		l.orderedTripods = append(l.orderedTripods, tri)
	}
}

func (l *Land) Execute(c *Ecall, ctx *Context, env *ChainEnv) error {
	Tripod, ok := l.tripodsMap[c.TripodName]
	if !ok {
		return TripodNotFound(c.TripodName)
	}
	ph := Tripod.GetTripodMeta()
	fn := ph.GetExec(c.ExecName)
	if fn == nil {
		return ExecNotFound(c.ExecName)
	}
	return fn(ctx, env)
}

func (l *Land) Query(c *Qcall, ctx *Context, env *ChainEnv) (interface{}, error) {
	Tripod, ok := l.tripodsMap[c.TripodName]
	if !ok {
		return nil, TripodNotFound(c.TripodName)
	}
	ph := Tripod.GetTripodMeta()
	qry := ph.GetQuery(c.QueryName)
	if qry == nil {
		return nil, QryNotFound(c.QueryName)
	}
	return qry(ctx, env, c.BlockHash)
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
