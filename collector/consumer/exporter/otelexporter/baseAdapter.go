package otelexporter

import "go.opentelemetry.io/otel/attribute"

type BaseAdapterManager struct {
	detailEntityAdapter   *Adapter
	aggEntityAdapter      *Adapter
	detailTopologyAdapter *Adapter
	aggTopologyAdapter    *Adapter
	traceToSpanAdapter    *Adapter
}

func createBaseAdapterManager(constLabels []attribute.KeyValue) *BaseAdapterManager {
	// TODO deal Error
	aggEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	detailEntityAdapter, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{entityInstanceMetricDicList, entityDetailMetricDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	aggTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(RemoveDstPodInfoForNonExternalAggTopology).
		build()

	detailTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(ReplaceDstIpOrDstPortByDNatIpAndDNatPortForDetailTopology).
		build()

	traceToSpanAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, SpanDicList}).
		withExtraLabels(spanProtocol, updateProtocolKey).
		withValueToLabels(traceSpanStatus, getTraceSpanStatusLabels).
		build()

	return &BaseAdapterManager{
		aggEntityAdapter:      aggEntityAdapter,
		detailEntityAdapter:   detailEntityAdapter,
		aggTopologyAdapter:    aggTopologyAdapter,
		detailTopologyAdapter: detailTopologyAdapter,
		traceToSpanAdapter:    traceToSpanAdapter,
	}
}
