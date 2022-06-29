package prometheusexporter

type MetricAggregationKind string

const (
	MAGaugeKind     MetricAggregationKind = "gauge"
	MACounterKind   MetricAggregationKind = "counter"
	MAHistogramKind MetricAggregationKind = "histogram"
)
