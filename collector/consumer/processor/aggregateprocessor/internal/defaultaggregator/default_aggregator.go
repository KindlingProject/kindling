package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal"
	"github.com/Kindling-project/kindling/collector/model"
	"sync"
)

// GaugeGroup
// Name:
//   labels:
//      GaugeGroup

type DefaultAggregator struct {
	recordersMap map[string]*valueRecorder
	// mut is only used to make sure the access to the recordersMap is thread-safe.
	// valueRecorder is responsible for its own thread-safe access.
	mut        sync.RWMutex
	aggKindMap map[string][]aggregatorKind
}

func NewDefaultAggregator(aggKindMap map[string][]string) *DefaultAggregator {
	ret := &DefaultAggregator{
		recordersMap: make(map[string]*valueRecorder),
		aggKindMap:   toAggKindMap(aggKindMap),
	}
	return ret
}

func (s *DefaultAggregator) Aggregate(g *model.GaugeGroup, f *internal.LabelFilter) {
	name := g.Name
	s.mut.RLock()
	recorder, ok := s.recordersMap[name]
	s.mut.RUnlock()
	// Won't enter this branch too many times, as the recordersMap
	// will become stable after running a period of time.
	if !ok {
		s.mut.Lock()
		// double check to avoid double writing
		recorder, ok = s.recordersMap[name]
		if !ok {
			recorder = newValueRecorder(name, g.Timestamp, s.aggKindMap)
			s.recordersMap[name] = recorder
		}
		s.mut.Unlock()
	}
	key := newLabelsKey(g.Labels, f)
	recorder.Record(key, g.Values)

	// First copy the model.GaugeGroup, then output the result directly.
	// Or first use intermediate struct, then generate the model.GaugeGroup.
}

func (s *DefaultAggregator) Dump() []*model.GaugeGroup {
	ret := make([]*model.GaugeGroup, 0)
	s.mut.RLock()
	for _, v := range s.recordersMap {
		ret = append(ret, v.dump()...)
		// Assume that the recordersMap will become stable after running a period of time.
		v.reset()
	}
	s.mut.RUnlock()
	return ret
}
