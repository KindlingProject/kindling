package processor

import (
	"github.com/dxsup/kindling-collector/consumer"
)

type Processor interface {
	consumer.Consumer
}
