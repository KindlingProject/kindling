package otelexporter

import "go.opentelemetry.io/otel/attribute"

type BaseAdapterManager struct {
	detailEntityAdapter   *metricAdapter
	aggEntityAdapter      *metricAdapter
	detailTopologyAdapter *metricAdapter
	aggTopologyAdapter    *metricAdapter
	traceToMetricAdapter  *metricAdapter
	traceToSpanAdapter    *metricAdapter
}

func createBaseAdapterManager(metricAggMap map[string]MetricAggregationKind, constLabels []attribute.KeyValue) *BaseAdapterManager {
	// TODO deal Error
	aggEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{},
		metricAggMap, true).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	detailEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{entityInstanceMetricDicList, entityDetailMetricDicList},
		metricAggMap, true).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	aggTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{},
		metricAggMap, false).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(RemoveDstPodInfoForNonExternalAggTopology).
		build()

	detailTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList},
		metricAggMap, false).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(ReplaceDstIpOrDstPortByDNatIpAndDNatPortForDetailTopology).
		build()

	traceToMetricAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList},
		// In traceToMetric, param `isServer` is not used
		metricAggMap, false).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withValueToLabels(traceStatus, getTraceStatusLabels).
		build()

	traceToSpanAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, SpanDicList},
		// In traceToSpan, param `isServer` is not used
		metricAggMap, false).
		withExtraLabels(spanProtocol, updateProtocolKey).
		withValueToLabels(traceSpanStatus, getTraceSpanStatusLabels).
		build()

	return &BaseAdapterManager{
		aggEntityAdapter:      aggEntityAdapter,
		detailEntityAdapter:   detailEntityAdapter,
		aggTopologyAdapter:    aggTopologyAdapter,
		detailTopologyAdapter: detailTopologyAdapter,
		traceToMetricAdapter:  traceToMetricAdapter,
		traceToSpanAdapter:    traceToSpanAdapter,
	}
}
