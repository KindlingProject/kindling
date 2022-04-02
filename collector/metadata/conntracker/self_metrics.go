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
	errorsTotal         = "kindling_telemetry_conntracker_errors_total"
	samplingRate        = "kindling_telemetry_conntracker_sampling_rate"
	throttlesTotal      = "kindling_telemetry_conntracker_throttles_total"
)

var (
	selfTelemetryOnce        sync.Once
	cacheSizeInstrument      metric.Int64GaugeObserver
	cacheMaxSizeInstrument   metric.Int64GaugeObserver
	operationTimesInstrument metric.Int64CounterObserver
	errorsTotalInstrument    metric.Int64CounterObserver
	samplingRateInstrument   metric.Int64GaugeObserver
	throttlesTotalInstrument metric.Int64CounterObserver
)

// To avoid getting statistics multiple times, we cache the result in a global variable.
// This can be done because the observation functions are executed in the order they were
// registered. See go.opentelemetry.io/otel/internal/metric/async/AsyncInstrumentState.runners.
var conntrackerStaticStates map[string]int64

func newSelfMetrics(meterProvider metric.MeterProvider, conntracker Conntracker) {
	selfTelemetryOnce.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		cacheSizeInstrument = meter.NewInt64GaugeObserver(cacheSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				conntrackerStaticStates = conntracker.GetStats()
				result.Observe(conntrackerStaticStates["state_size"], attribute.String("type", "general"))
				result.Observe(conntrackerStaticStates["orphan_size"], attribute.String("type", "orphan"))
			})
		cacheMaxSizeInstrument = meter.NewInt64GaugeObserver(cacheMaxSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["cache_max_size"])
			})
		operationTimesInstrument = meter.NewInt64CounterObserver(operationTimesTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["registers_total"], attribute.String("op", "add"))
				result.Observe(conntrackerStaticStates["registers_dropped"], attribute.String("op", "drop"))
				result.Observe(conntrackerStaticStates["unregisters_total"], attribute.String("op", "remove"))
				result.Observe(conntrackerStaticStates["gets_total"], attribute.String("op", "get"))
				result.Observe(conntrackerStaticStates["evicts_total"], attribute.String("op", "evict"))
			})
		errorsTotalInstrument = meter.NewInt64CounterObserver(errorsTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["enobufs"], attribute.String("type", "enobuf"))
				result.Observe(conntrackerStaticStates["read_errors"], attribute.String("type", "read_errors"))
				result.Observe(conntrackerStaticStates["msg_errors"], attribute.String("type", "msg_errors"))
			})
		samplingRateInstrument = meter.NewInt64GaugeObserver(samplingRate,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["sampling_pct"])
			})
		throttlesTotalInstrument = meter.NewInt64CounterObserver(throttlesTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["throttles"])
			})
		// Suppress warnings of unused variables
		_ = cacheSizeInstrument
		_ = cacheMaxSizeInstrument
		_ = operationTimesInstrument
		_ = errorsTotalInstrument
		_ = samplingRateInstrument
		_ = throttlesTotalInstrument
	})
}
