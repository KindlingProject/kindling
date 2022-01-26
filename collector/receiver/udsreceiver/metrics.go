package udsreceiver

import "go.opentelemetry.io/otel/metric"

var eventReceivedMetric = "kindling_telemetry_event_received_total"

type selfMetrics struct {
	eventSentCounter metric.Int64Counter
}

func NewSelfMetrics(meterProvider metric.MeterProvider) *selfMetrics {
	return &selfMetrics{
		eventSentCounter: metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(eventReceivedMetric),
	}
}
