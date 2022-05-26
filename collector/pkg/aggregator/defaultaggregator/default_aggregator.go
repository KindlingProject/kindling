package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"sync"
)

// DataGroup
// Name:
//   labels:
//      DataGroup

type DefaultAggregator struct {
	recordersMap sync.Map
	// mut is only used to make sure Dump the recorder is thread-safe.
	// valueRecorder is responsible for its own thread-safe access.
	mut    sync.RWMutex
	config *AggregatedConfig
}

func NewDefaultAggregator(config *AggregatedConfig) *DefaultAggregator {
	ret := &DefaultAggregator{
		recordersMap: sync.Map{},
		config:       config,
	}
	return ret
}

func (s *DefaultAggregator) Aggregate(g *model.DataGroup, selectors *aggregator.LabelSelectors) {
	name := g.Name
	s.mut.RLock()
	defer s.mut.RUnlock()
	recorder, ok := s.recordersMap.Load(name)
	// Won't enter this branch too many times, as the recordersMap
	// will become stable after running a period of time.
	if !ok {
		// double check to avoid double writing
		recorder, _ = s.recordersMap.LoadOrStore(name, newValueRecorder(name, s.config.KindMap))
	}
	key := selectors.GetLabelKeys(g.Labels)
	recorder.(*valueRecorder).Record(key, g.Metrics, g.Timestamp)
}

func (s *DefaultAggregator) Dump() []*model.DataGroup {
	ret := make([]*model.DataGroup, 0)
	s.mut.Lock()
	defer s.mut.Unlock()
	s.recordersMap.Range(func(_, value interface{}) bool {
		recorder := value.(*valueRecorder)
		ret = append(ret, recorder.dump()...)
		// Assume that the recordersMap will become stable after running a period of time.
		recorder.reset()
		return true
	})
	return ret
}

func (s *DefaultAggregator) DumpSingle(metricGroupName string) []*model.DataGroup {
	s.mut.Lock()
	defer s.mut.Unlock()
	if v, ok := s.recordersMap.Load(metricGroupName); ok {
		vr := v.(*valueRecorder)
		result := vr.dump()
		vr.reset()
		return result
	}
	return nil
}

func (s *DefaultAggregator) CheckExist(metricGroupName string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	_, result := s.recordersMap.Load(metricGroupName)
	return result
}
