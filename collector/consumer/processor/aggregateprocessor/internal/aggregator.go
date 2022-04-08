package internal

import (
	"github.com/Kindling-project/kindling/collector/model"
)

type Aggregator interface {
	Aggregate(g *model.GaugeGroup, f *LabelFilter)
	Dump() []*model.GaugeGroup
}

// Exclude labels
// General: src_port
// HTTP: HttpUrl, HttpApmTraceId, HttpRequestPayload, HttpResponsePayload
// DNS: DnsId, DnsIp
// Sql: Sql, SqlErrMsg
// RedisErrMsg: RedisErrMsg
// Kafka: KafkaApi, KafkaVersion, KafkaCorrelationId

// LabelFilter is used to filter the labels which should not be used as keys when run aggregation
type LabelFilter struct {
	set map[string]bool
}

func NewLabelFilter(labels ...string) *LabelFilter {
	ret := &LabelFilter{set: make(map[string]bool, len(labels))}
	for _, label := range labels {
		ret.set[label] = true
	}
	return ret
}

func (f *LabelFilter) Filter(label string) bool {
	_, ok := f.set[label]
	return ok
}
