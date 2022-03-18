package conntracker

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

const (
	cacheSizeMetric     = "kindling_telemetry_conntracker_cache_size"
	cacheMaxSizeMetric  = "kindling_telemetry_conntracker_cache_max_size"
	operationTimesTotal = "kindling_telemetry_conntracker_operation_times_total"
)

var (
	selfTelemetryOnce        sync.Once
	cacheSizeInstrument      metric.Int64GaugeObserver
	cacheMaxSizeInstrument   metric.Int64GaugeObserver
	operationTimesInstrument metric.Int64GaugeObserver
)

func newSelfMetrics(meterProvider metric.MeterProvider, conntracker *Conntracker) {
	selfTelemetryOnce.Do(func() {
		cacheSizeInstrument = metric.Must(meterProvider.Meter("kindling")).NewInt64GaugeObserver(cacheSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(conntracker.cache.Len()))
			})
		cacheMaxSizeInstrument = metric.Must(meterProvider.Meter("kindling")).NewInt64GaugeObserver(cacheMaxSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(conntracker.maxCacheSize))
			})
		operationTimesInstrument = metric.Must(meterProvider.Meter("kindling")).NewInt64GaugeObserver(operationTimesTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntracker.cache.stats.add, attribute.String("op", "add"))
				result.Observe(conntracker.cache.stats.gets, attribute.String("op", "get"))
				result.Observe(conntracker.cache.stats.remove, attribute.String("op", "remove"))
				result.Observe(conntracker.cache.stats.evicts, attribute.String("op", "evict"))
			})
		// Suppress warnings of unused variables
		_ = cacheSizeInstrument
		_ = cacheMaxSizeInstrument
		_ = operationTimesInstrument
	})
}
