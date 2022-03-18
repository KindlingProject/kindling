package network

import "go.opentelemetry.io/otel/metric"

var netanalyzerMessagePairMetric = "kindling_telemetry_netanalyer_messagepair_size"
var netanalyzerParsedRequestMetric = "kindling_telemetry_netanalyer_parsedrequest_total"

type selfMetrics struct {
	netanalyzerMessagePairSize    metric.Int64UpDownCounter
	netanalyzerParsedRequestTotal metric.Int64Counter
}

func NewSelfMetrics(meterProvider metric.MeterProvider) *selfMetrics {
	return &selfMetrics{
		netanalyzerMessagePairSize:    metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounter(netanalyzerMessagePairMetric),
		netanalyzerParsedRequestTotal: metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(netanalyzerParsedRequestMetric),
	}
}
