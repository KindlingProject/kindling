package otelexporter

import "go.opentelemetry.io/otel/attribute"

type BaseAdapterManager struct {
	detailEntityAdapter   *metricAdapter
	aggEntityAdapter      *metricAdapter
	detailTopologyAdapter *metricAdapter
	aggTopologyAdapter    *metricAdapter
}

func createBaseAdapterManager(metricAggMap map[string]MetricAggregationKind, constLabels []attribute.KeyValue) *BaseAdapterManager {
	// TODO deal Error
	aggEntityAdapter, _ := newMetricAdapterBuilder(entityMetricDicList, [][]dictionary{instanceMetricDicList},
		metricAggMap, true).
		withProtocols(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	return &BaseAdapterManager{
		aggEntityAdapter: aggEntityAdapter,
	}
}
