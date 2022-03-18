package network

import (
	"sync"

	"go.opentelemetry.io/otel/metric"
)

var once sync.Once
var netanalyzerMessagePairSize metric.Int64UpDownCounter
var netanalyzerParsedRequestTotal metric.Int64Counter

const (
	netanalyzerMessagePairMetric   = "kindling_telemetry_netanalyer_messagepair_size"
	netanalyzerParsedRequestMetric = "kindling_telemetry_netanalyer_parsedrequest_total"
)

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		netanalyzerMessagePairSize = metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounter(netanalyzerMessagePairMetric)
		netanalyzerParsedRequestTotal = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(netanalyzerParsedRequestMetric)
	})
}
