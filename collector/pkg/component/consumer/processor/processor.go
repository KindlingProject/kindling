package processor

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
)

type Processor interface {
	consumer.Consumer
}
