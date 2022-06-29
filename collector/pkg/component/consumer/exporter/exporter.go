package exporter

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
)

type Exporter interface {
	consumer.Consumer
}
