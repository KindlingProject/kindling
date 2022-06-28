package aggregator

import (
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type Aggregator interface {
	Aggregate(g *model.DataGroup, s *LabelSelectors)
	Dump() []*model.DataGroup
}
