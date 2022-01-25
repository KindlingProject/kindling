package processor

import (
	"github.com/Kindling-project/kindling/collector/consumer"
)

type Processor interface {
	consumer.Consumer
}
