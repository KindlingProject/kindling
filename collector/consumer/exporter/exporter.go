package exporter

import (
	"github.com/dxsup/kindling-collector/consumer"
)

type Exporter interface {
	consumer.Consumer
}
