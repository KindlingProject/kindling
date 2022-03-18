package udsreceiver

import (
	"go.opentelemetry.io/otel/metric"
	"sync"
)

var once sync.Once
var globalEventSentCounter metric.Int64Counter

const eventReceivedMetric = "kindling_telemetry_udsreceiver_events_total"

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		globalEventSentCounter = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(eventReceivedMetric)
	})
}
