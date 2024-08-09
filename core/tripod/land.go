package tripod

import (
	. "github.com/yu-org/yu/common/yerror"
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

func (l *Land) GetTripod(name string) *Tripod {
	return l.TripodsMap[name]
}

func (l *Land) GetWriting(tripodName, wrName string) (Writing, error) {
	tripod, ok := l.TripodsMap[tripodName]
	if !ok {
		return nil, TripodNotFound(tripodName)
	}
	fn := tripod.GetWriting(wrName)
	if fn == nil {
		return nil, WritingNotFound(wrName)
	}
	return fn, nil
}

func (l *Land) GetReading(tripodName, rdName string) (Reading, error) {
	tri, ok := l.TripodsMap[tripodName]
	if !ok {
		return nil, TripodNotFound(tripodName)
	}
	rd := tri.GetReading(rdName)
	if rd == nil {
		return nil, ReadingNotFound(rdName)
	}
	return rd, nil
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
