package tripod

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/context"
	"github.com/yu-org/yu/core/tripod/dev"
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

func (l *Land) GetExecLei(c *Ecall) (dev.Execution, uint64, error) {
	tripod, ok := l.tripodsMap[c.TripodName]
	if !ok {
		return nil, 0, TripodNotFound(c.TripodName)
	}
	ph := tripod.GetTripodMeta()
	fn, lei := ph.GetExec(c.ExecName)
	if fn == nil {
		return nil, 0, ExecNotFound(c.ExecName)
	}
	return fn, lei, nil
}

func (l *Land) Query(c *Qcall, ctx *Context) (interface{}, error) {
	tripod, ok := l.tripodsMap[c.TripodName]
	if !ok {
		return nil, TripodNotFound(c.TripodName)
	}
	ph := tripod.GetTripodMeta()
	qry := ph.GetQuery(c.QueryName)
	if qry == nil {
		return nil, QryNotFound(c.QueryName)
	}
	return qry(ctx, c.BlockHash)
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