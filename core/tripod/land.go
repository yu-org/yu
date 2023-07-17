package tripod

import (
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/tripod/dev"
)

type Land struct {
	OrderedTripods []*Tripod
	// Key: the Name of Tripod
	TripodsMap map[string]*Tripod
}

func NewLand() *Land {
	return &Land{
		TripodsMap:     make(map[string]*Tripod),
		OrderedTripods: make([]*Tripod, 0),
	}
}

func (l *Land) SetTripods(Tripods ...*Tripod) {
	for _, tri := range Tripods {
		triName := tri.Name()
		l.TripodsMap[triName] = tri

		l.OrderedTripods = append(l.OrderedTripods, tri)
	}
}

func (l *Land) GetTripodInstance(name string) interface{} {
	if tri, ok := l.TripodsMap[name]; ok {
		return tri.Instance
	}
	return nil
}

func (l *Land) GetWriting(c *WrCall) (Writing, error) {
	tripod, ok := l.TripodsMap[c.TripodName]
	if !ok {
		return nil, TripodNotFound(c.TripodName)
	}
	fn := tripod.GetWriting(c.WritingName)
	if fn == nil {
		return nil, WritingNotFound(c.WritingName)
	}
	return fn, nil
}

func (l *Land) Read(c *RdCall, ctx *ReadContext) error {
	tri, ok := l.TripodsMap[c.TripodName]
	if !ok {
		return TripodNotFound(c.TripodName)
	}
	rd := tri.GetReading(c.ReadingName)
	if rd == nil {
		return ReadingNotFound(c.ReadingName)
	}
	return rd(ctx)
}

func (l *Land) RangeMap(fn func(string, *Tripod) error) error {
	for name, tri := range l.TripodsMap {
		err := fn(name, tri)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Land) RangeList(fn func(*Tripod) error) error {
	for _, tri := range l.OrderedTripods {
		err := fn(tri)
		if err != nil {
			return err
		}
	}
	return nil
}
