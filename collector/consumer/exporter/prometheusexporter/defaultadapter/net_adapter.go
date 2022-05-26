package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
)

type NetMetricGroupAdapter struct {
	*NetAdapterManager
	*NetAdapterConfig
}

type NetAdapterConfig struct {
	StoreTraceAsMetric bool
	StoreTraceAsSpan   bool
	StorePodDetail     bool
	StoreExternalSrcIP bool
}

func (n *NetMetricGroupAdapter) Adapt(metricGroup *model.DataGroup) ([]*AdaptedResult, error) {
	switch metricGroup.Name {
	case constnames.AggregatedNetRequestMetricGroup:
		return n.dealWithPreAggMetricGroups(metricGroup), nil
	case constnames.SingleNetRequestMetricGroup:
		return n.dealWithSingleMetricGroup(metricGroup), nil
	default:
		return nil, nil
	}
}

func (n *NetMetricGroupAdapter) dealWithSingleMetricGroup(metricGroup *model.DataGroup) []*AdaptedResult {
	requestTotalTime, ok := metricGroup.GetMetric(constvalues.RequestTotalTime)
	if !ok {
		return nil
	}
	results := make([]*AdaptedResult, 0, 2)
	if n.StoreTraceAsMetric {
		labels, free := n.traceToMetricAdapter.transform(metricGroup)
		results = append(results, &AdaptedResult{
			ResultType:   Metric,
			Metrics:      []*model.Metric{model.NewIntMetric(constnames.TraceAsMetric, requestTotalTime.GetInt().Value)},
			AttrsMap:     labels,
			Timestamp:    metricGroup.Timestamp,
			FreeAttrsMap: free,
		})
	}

	return results
}

func (n *NetMetricGroupAdapter) dealWithPreAggMetricGroups(metricGroup *model.DataGroup) []*AdaptedResult {
	results := make([]*AdaptedResult, 0, 4)
	isServer := metricGroup.Labels.GetBoolValue(constlabels.IsServer)
	srcNamespace := metricGroup.Labels.GetStringValue(constlabels.SrcNamespace)
	if n.StoreExternalSrcIP && srcNamespace == constlabels.ExternalClusterNamespace && isServer {
		externalAdapterCache := n.detailTopologyAdapter
		externalTopology := n.createNetMetricResults(metricGroup, externalAdapterCache)
		results = append(results, externalTopology...)
	}

	var metricAdapterCache [2]*LabelConverter
	if n.StorePodDetail {
		if isServer {
			metricAdapterCache = n.detailEntityAdapter
		} else {
			metricAdapterCache = n.detailTopologyAdapter
		}
	} else {
		if isServer {
			metricAdapterCache = n.aggEntityAdapter
		} else {
			metricAdapterCache = n.aggTopologyAdapter
		}
	}
	metrics := n.createNetMetricResults(metricGroup, metricAdapterCache)
	return append(results, metrics...)
}

func (n *NetMetricGroupAdapter) createNetMetricResults(metricGroup *model.DataGroup, adapter [2]*LabelConverter) (tmpResults []*AdaptedResult) {
	values := metricGroup.Metrics
	isServer := metricGroup.Labels.GetBoolValue(constlabels.IsServer)
	metricsExceptRequestCount := make([]*model.Metric, 0, len(values))
	requestCount := make([]*model.Metric, 0, 1)
	for _, metric := range metricGroup.Metrics {
		if metric.Name != constvalues.RequestCount {
			metricsExceptRequestCount = append(metricsExceptRequestCount, model.NewIntMetric(
				constnames.ToKindlingNetMetricName(metric.Name, isServer),
				metric.GetInt().Value))
		} else {
			requestCount = append(requestCount, model.NewIntMetric(
				constnames.ToKindlingNetMetricName(metric.Name, isServer),
				metric.GetInt().Value))
		}
	}
	attrsCommon, free := adapter[0].transform(metricGroup)
	tmpResults = make([]*AdaptedResult, 0, 2)
	if len(metricsExceptRequestCount) > 0 {
		// for request count
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType:   Metric,
			AttrsMap:     attrsCommon,
			Metrics:      metricsExceptRequestCount,
			Timestamp:    metricGroup.Timestamp,
			FreeAttrsMap: free,
		})
	}
	if len(requestCount) > 0 {
		attrsWithSlow, free := adapter[1].transform(metricGroup)
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType:   Metric,
			AttrsMap:     attrsWithSlow,
			Metrics:      requestCount,
			Timestamp:    metricGroup.Timestamp,
			FreeAttrsMap: free,
		})
	}
	return
}

type NetAdapterManager struct {
	detailEntityAdapter   [2]*LabelConverter
	aggEntityAdapter      [2]*LabelConverter
	detailTopologyAdapter [2]*LabelConverter
	aggTopologyAdapter    [2]*LabelConverter
	traceToSpanAdapter    *LabelConverter
	traceToMetricAdapter  *LabelConverter
}

func createNetAdapterManager(constLabels []attribute.KeyValue) *NetAdapterManager {
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
		withAdjust(removeDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapterWithIsSlow, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList, isSlowDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(replaceDstIpOrDstPortByDNat()).
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
		withAdjust(removeDstPodInfoForNonExternal()).
		withConstLabels(constLabels).
		build()

	detailTopologyAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
		withExtraLabels(topologyProtocol, updateProtocolKey).
		withAdjust(replaceDstIpOrDstPortByDNat()).
		withConstLabels(constLabels).
		build()

	traceToSpanAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, SpanDicList, dNatDicList}).
		withExtraLabels(spanProtocol, updateProtocolKey).
		withValueToLabels(traceSpanStatus, getTraceSpanStatusLabels).
		withConstLabels(constLabels).
		build()

	traceToMetricAdapter, _ := newAdapterBuilder(topologyMetricDicList,
		[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList, dNatDicList}).
		withExtraLabels(entityProtocol, updateProtocolKey).
		withValueToLabels(traceStatus, getTraceStatusLabels).
		withConstLabels(constLabels).
		build()

	return &NetAdapterManager{
		aggEntityAdapter:      [2]*LabelConverter{aggEntityAdapter, aggEntityAdapterWithIsSlow},
		detailEntityAdapter:   [2]*LabelConverter{detailEntityAdapter, detailEntityAdapterWithIsSlow},
		aggTopologyAdapter:    [2]*LabelConverter{aggTopologyAdapter, aggTopologyAdapterWithIsSlow},
		detailTopologyAdapter: [2]*LabelConverter{detailTopologyAdapter, detailTopologyAdapterWithIsSlow},
		traceToSpanAdapter:    traceToSpanAdapter,
		traceToMetricAdapter:  traceToMetricAdapter,
	}
}

func NewNetAdapter(
	customLabels []attribute.KeyValue,
	config *NetAdapterConfig,
) *NetMetricGroupAdapter {
	return &NetMetricGroupAdapter{
		NetAdapterManager: createNetAdapterManager(customLabels),
		NetAdapterConfig:  config,
	}
}
