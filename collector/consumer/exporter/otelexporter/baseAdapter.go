package otelexporter

import "go.opentelemetry.io/otel/attribute"

type BaseAdapterManager struct {
	detailEntityAdapter   [2]*Adapter
	aggEntityAdapter      [2]*Adapter
	detailTopologyAdapter [2]*Adapter
	aggTopologyAdapter    [2]*Adapter
	traceToSpanAdapter    *Adapter
	traceToMetricAdapter  *Adapter
}

func createBaseAdapterManager(constLabels []attribute.KeyValue) *BaseAdapterManager {
	// TODO deal Error
	aggEntityAdapterWithIsSlow, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{isSlowDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	detailEntityAdapterWithIsSlow, _ := newAdapterBuilder(entityMetricDicList,
		[][]dictionary{entityInstanceMetricDicList, entityDetailMetricDicList, isSlowDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withConstLabels(constLabels).
		build()

	aggTopologyAdapterWithIsSlow, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{isSlowDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(RemoveDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapterWithIsSlow, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList, isSlowDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(ReplaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

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
		withAdjust(RemoveDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(ReplaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	traceToSpanAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, SpanDicList}).
		withExtraLabels(spanProtocol, updateProtocolKey).
		withValueToLabels(traceSpanStatus, getTraceSpanStatusLabels).
		//withAdjust(ReplaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	traceToMetricAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withValueToLabels(traceStatus, getTraceStatusLabels).
		//withAdjust(ReplaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	return &BaseAdapterManager{
		aggEntityAdapter:      [2]*Adapter{aggEntityAdapter, aggEntityAdapterWithIsSlow},
		detailEntityAdapter:   [2]*Adapter{detailEntityAdapter, detailEntityAdapterWithIsSlow},
		aggTopologyAdapter:    [2]*Adapter{aggTopologyAdapter, aggTopologyAdapterWithIsSlow},
		detailTopologyAdapter: [2]*Adapter{detailTopologyAdapter, detailTopologyAdapterWithIsSlow},
		traceToSpanAdapter:    traceToSpanAdapter,
		traceToMetricAdapter:  traceToMetricAdapter,
	}
}
