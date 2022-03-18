package network

import "go.opentelemetry.io/otel/metric"

var networkRequestSizeMetric = "kindling_telemetry_network_request_size"
var networkProtocolTotalMetric = "kindling_telemetry_network_protocol_total"

type selfMetrics struct {
	networkRequestSize   metric.Int64UpDownCounter
	networkProtocolTotal metric.Int64Counter
}

func NewSelfMetrics(meterProvider metric.MeterProvider) *selfMetrics {
	return &selfMetrics{
		networkRequestSize:   metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounter(networkRequestSizeMetric),
		networkProtocolTotal: metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(networkProtocolTotalMetric),
	}
}
