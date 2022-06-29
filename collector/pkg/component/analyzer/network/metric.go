package network

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	netanalyzerMessagePairMetric   = "kindling_telemetry_netanalyer_messagepair_size"
	netanalyzerParsedRequestMetric = "kindling_telemetry_netanalyer_parsedrequest_total"
)

var (
	selfTelemetryOnce                    sync.Once
	netanalyzerMessagePairSizeInstrument metric.Int64GaugeObserver
	netanalyzerParsedRequestTotal        metric.Int64Counter
)

func newSelfMetrics(meterProvider metric.MeterProvider, na *NetworkAnalyzer) {
	selfTelemetryOnce.Do(func() {
		netanalyzerMessagePairSizeInstrument = metric.Must(meterProvider.Meter("kindling")).NewInt64GaugeObserver(netanalyzerMessagePairMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(na.tcpMessagePairSize, attribute.String("type", "tcp"))
				result.Observe(na.udpMessagePairSize, attribute.String("type", "udp"))
			})
		netanalyzerParsedRequestTotal = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(netanalyzerParsedRequestMetric)
		// Suppress warnings of unused variables
		_ = netanalyzerMessagePairSizeInstrument
	})
}
