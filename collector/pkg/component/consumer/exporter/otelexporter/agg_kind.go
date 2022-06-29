package otelexporter

type MetricAggregationKind string

const (
	MAGaugeKind     MetricAggregationKind = "gauge"
	MACounterKind   MetricAggregationKind = "counter"
	MAHistogramKind MetricAggregationKind = "histogram"
)
