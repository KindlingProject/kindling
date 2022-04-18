package internal

import (
	"github.com/Kindling-project/kindling/collector/model"
)

type Aggregator interface {
	Aggregate(g *model.GaugeGroup, s *LabelSelectors)
	Dump() []*model.GaugeGroup
}
