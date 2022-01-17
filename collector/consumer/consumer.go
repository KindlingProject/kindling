package consumer

import "github.com/dxsup/kindling-collector/model"

type Consumer interface {
	Consume(gaugeGroup *model.GaugeGroup) error
}
