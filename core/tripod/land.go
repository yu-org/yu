package tripod

import (
	"fmt"
	"strings"

	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/tripod/dev"
)

type Land struct {
	orderedTripods []*Tripod
	// Key: the Name of Tripod
	tripodsMap map[string]*Tripod

	// Key: topic::tripodName
	topicTripods        map[string]*Tripod
	orderedTopicTripods []string

	bronzes map[string]*Bronze
}

func NewLand() *Land {
	return &Land{
		tripodsMap:          make(map[string]*Tripod),
		orderedTripods:      make([]*Tripod, 0),
		topicTripods:        make(map[string]*Tripod),
		orderedTopicTripods: make([]string, 0),
		bronzes:             make(map[string]*Bronze),
	}
}

func (l *Land) RegisterBronzes(bronzes ...*Bronze) {
	for _, bronze := range bronzes {
		l.bronzes[bronze.Name()] = bronze
	}
}

func (l *Land) RegisterTripods(tripods ...*Tripod) {
	for _, tri := range tripods {
		triName := tri.Name()

		l.tripodsMap[triName] = tri
		l.orderedTripods = append(l.orderedTripods, tri)

		for topic := range tri.topicWritings {
			l.registerTopicTripod(topic, tri)
		}
	}
}

const topicTripodKeySep = "::"

func makeTopicTripodKey(topic, tripodName string) string {
	return fmt.Sprintf("%s%s%s", topic, topicTripodKeySep, tripodName)
}

func (l *Land) registerTopicTripod(topic string, tri *Tripod) {
	key := makeTopicTripodKey(topic, tri.Name())
	if _, exists := l.topicTripods[key]; !exists {
		l.orderedTopicTripods = append(l.orderedTopicTripods, key)
	}
	l.topicTripods[key] = tri
}

func splitTopicTripodKey(key string) (topic, tripodName string) {
	parts := strings.SplitN(key, topicTripodKeySep, 2)
	if len(parts) != 2 {
		return key, ""
	}
	return parts[0], parts[1]
}

type TopicTripod struct {
	Topic      string
	TripodName string
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

func (l *Land) GetTopicWriting(tripodName, ewName, topic string) (TopicWriting, error) {
	key := makeTopicTripodKey(topic, tripodName)
	tripod, ok := l.topicTripods[key]
	if !ok {
		if _, exist := l.tripodsMap[tripodName]; !exist {
			return nil, TripodNotFound(tripodName)
		}
		return nil, TopicWritingNotFound(topic)
	}
	ew := tripod.GetTopicWriting(topic)
	if ew == nil {
		return nil, TopicWritingNotFound(topic)
	}
	_ = ewName
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

func (l *Land) TopicTripods() map[string]*Tripod {
	result := make(map[string]*Tripod, len(l.topicTripods))
	for key, tri := range l.topicTripods {
		result[key] = tri
	}
	return result
}

func (l *Land) OrderedTopicTripods() []TopicTripod {
	result := make([]TopicTripod, 0, len(l.orderedTopicTripods))
	for _, key := range l.orderedTopicTripods {
		topic, tripodName := splitTopicTripodKey(key)
		result = append(result, TopicTripod{
			Topic:      topic,
			TripodName: tripodName,
		})
	}
	return result
}

func (l *Land) TopicNames() []string {
	names := make([]string, 0, len(l.orderedTopicTripods))
	seen := make(map[string]struct{})
	for _, key := range l.orderedTopicTripods {
		topic, _ := splitTopicTripodKey(key)
		if _, exists := seen[topic]; exists {
			continue
		}
		seen[topic] = struct{}{}
		names = append(names, topic)
	}
	return names
}
