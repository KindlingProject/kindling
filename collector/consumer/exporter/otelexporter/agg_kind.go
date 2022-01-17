package otelexporter

type MetricAggregationKind int32

const (
	MAGaugeKind MetricAggregationKind = iota
	MACounterKind
	MAHistogramKind
)
