package exporter

import (
	"github.com/Kindling-project/kindling/collector/consumer"
)

type Exporter interface {
	consumer.Consumer
}
