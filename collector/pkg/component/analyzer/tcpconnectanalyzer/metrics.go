package tcpconnectanalyzer

import (
	"context"
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/tcpconnectanalyzer/internal"
	"go.opentelemetry.io/otel/metric"
)

var once sync.Once

const mapSizeMetric = "kindling_telemetry_tcpconnectanalyzer_map_size"

func newSelfMetrics(meterProvider metric.MeterProvider, monitor *internal.ConnectMonitor) {
	once.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		meter.NewInt64GaugeObserver(mapSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(monitor.GetMapSize()))
			})
	})
}
