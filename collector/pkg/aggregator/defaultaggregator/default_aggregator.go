package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"sync"
)

// GaugeGroup
// Name:
//   labels:
//      GaugeGroup

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

func (s *DefaultAggregator) Aggregate(g *model.GaugeGroup, selectors *aggregator.LabelSelectors) {
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
	recorder.(*valueRecorder).Record(key, g.Values, g.Timestamp)
}

func (s *DefaultAggregator) Dump() []*model.GaugeGroup {
	ret := make([]*model.GaugeGroup, 0)
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

func (s *DefaultAggregator) DumpSingle(gaugeGroupName string) []*model.GaugeGroup {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if v, ok := s.recordersMap.Load(gaugeGroupName); ok {
		vr := v.(*valueRecorder)
		result := vr.dump()
		vr.reset()
		return result
	}
	return nil
}

func (s *DefaultAggregator) CheckExist(gaugeGroupName string) bool {
	s.mut.Lock()
	defer s.mut.Unlock()
	_, result := s.recordersMap.Load(gaugeGroupName)
	return result
}
