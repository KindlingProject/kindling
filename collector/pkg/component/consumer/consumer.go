package consumer

import "github.com/Kindling-project/kindling/collector/pkg/model"

type Consumer interface {
	Consume(dataGroup *model.DataGroup) error
}
