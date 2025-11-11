package tripod

import (
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/tripod/dev"
)

type Land struct {
	orderedTripods []*Tripod
	// Key: the Name of Tripod
	tripodsMap map[string]*Tripod

	bronzes map[string]*Bronze
}

func NewLand() *Land {
	return &Land{
		tripodsMap:     make(map[string]*Tripod),
		orderedTripods: make([]*Tripod, 0),
		bronzes:        make(map[string]*Bronze),
	}
}

func (l *Land) SetBronzes(bronzes ...*Bronze) {
	for _, bronze := range bronzes {
		l.bronzes[bronze.Name()] = bronze
	}
}

func (l *Land) SetTripods(tripods ...*Tripod) {
	for _, tri := range tripods {
		triName := tri.Name()

		l.tripodsMap[triName] = tri
		l.orderedTripods = append(l.orderedTripods, tri)
	}
}

func (l *Land) GetTripodInstance(name string) interface{} {
	if tri, ok := l.tripodsMap[name]; ok {
		return tri.Instance
	}
	return nil
}

func (l *Land) GetTripod(name string) *Tripod {
	return l.tripodsMap[name]
}

func (l *Land) GetWriting(tripodName, wrName string) (Writing, error) {
	tripod, ok := l.tripodsMap[tripodName]
	if !ok {
		return nil, TripodNotFound(tripodName)
	}
	fn := tripod.GetWriting(wrName)
	if fn == nil {
		return nil, WritingNotFound(wrName)
	}
	return fn, nil
}

func (l *Land) GetTopicWriting(tripodName, ewName string) (TopicWriting, error) {
	tripod, ok := l.tripodsMap[tripodName]
	if !ok {
		return nil, TripodNotFound(tripodName)
	}
	ew := tripod.GetTopicWriting(ewName)
	if ew == nil {
		return nil, TopicWritingNotFound(ewName)
	}
	return ew, nil
}

func (l *Land) GetReading(tripodName, rdName string) (Reading, error) {
	tri, ok := l.tripodsMap[tripodName]
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
	for name, tri := range l.tripodsMap {
		err := fn(name, tri)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *Land) RangeList(fn func(*Tripod) error) error {
	for _, tri := range l.orderedTripods {
		err := fn(tri)
		if err != nil {
			return err
		}
	}
	return nil
}
