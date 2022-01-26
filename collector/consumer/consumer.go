package consumer

import "github.com/Kindling-project/kindling/collector/model"

type Consumer interface {
	Consume(gaugeGroup *model.GaugeGroup) error
}
