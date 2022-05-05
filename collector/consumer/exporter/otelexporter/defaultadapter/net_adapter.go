package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
)

type NetGaugeGroupAdapter struct {
	*NetAdapterManager
	*NetAdapterConfig
}

type NetAdapterConfig struct {
	StoreTraceAsMetric bool
	StoreTraceAsSpan   bool
	StorePodDetail     bool
	StoreExternalSrcIP bool
}

func (n *NetGaugeGroupAdapter) Adapt(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	switch gaugeGroup.Name {
	case constnames.AggregatedNetRequestGaugeGroup:
		return n.dealWithPreAggGaugeGroups(gaugeGroup), nil
	case constnames.SingleNetRequestGaugeGroup:
		return n.dealWithSingleGaugeGroup(gaugeGroup), nil
	default:
		return nil, nil
	}
}

func (n *NetGaugeGroupAdapter) dealWithSingleGaugeGroup(gaugeGroup *model.GaugeGroup) []*AdaptedResult {
	requestTotalTime, ok := gaugeGroup.GetGauge(constvalues.RequestTotalTime)
	if !ok {
		return nil
	}
	results := make([]*AdaptedResult, 0, 2)
	if n.StoreTraceAsSpan {
		attrs, free := n.traceToSpanAdapter.convert(gaugeGroup)
		results = append(results, &AdaptedResult{
			ResultType:    Trace,
			AttrsList:     attrs,
			Gauges:        []*model.Gauge{requestTotalTime},
			Timestamp:     gaugeGroup.Timestamp,
			FreeAttrsList: free,
		})
	}
	if n.StoreTraceAsMetric {
		labels, free := n.traceToMetricAdapter.transform(gaugeGroup)
		results = append(results, &AdaptedResult{
			ResultType: Metric,
			AttrsList:  nil,
			Gauges: []*model.Gauge{{
				Name:  constnames.TraceAsMetric,
				Value: requestTotalTime.Value,
			}},
			AttrsMap:     labels,
			Timestamp:    gaugeGroup.Timestamp,
			FreeAttrsMap: free,
		})
	}

	return results
}

func (n *NetGaugeGroupAdapter) dealWithPreAggGaugeGroups(gaugeGroup *model.GaugeGroup) []*AdaptedResult {
	results := make([]*AdaptedResult, 0, 4)
	isServer := gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)
	srcNamespace := gaugeGroup.Labels.GetStringValue(constlabels.SrcNamespace)
	if n.StoreExternalSrcIP && srcNamespace == constlabels.ExternalClusterNamespace && isServer {
		externalAdapterCache := n.detailTopologyAdapter
		externalTopology := n.createNetMetricResults(gaugeGroup, externalAdapterCache)
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
	metrics := n.createNetMetricResults(gaugeGroup, metricAdapterCache)
	return append(results, metrics...)
}

func (n *NetGaugeGroupAdapter) createNetMetricResults(gaugeGroup *model.GaugeGroup, adapter [2]*LabelConverter) (tmpResults []*AdaptedResult) {
	values := gaugeGroup.Values
	isServer := gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)
	gaugesExceptRequestCount := make([]*model.Gauge, 0, len(values))
	requestCount := make([]*model.Gauge, 0, 1)
	for _, gauge := range gaugeGroup.Values {
		if gauge.Name != constvalues.RequestCount {
			gaugesExceptRequestCount = append(gaugesExceptRequestCount, &model.Gauge{
				Name:  constnames.ToKindlingNetMetricName(gauge.Name, isServer),
				Value: gauge.Value,
			})
		} else {
			requestCount = append(requestCount, &model.Gauge{
				Name:  constnames.ToKindlingNetMetricName(gauge.Name, isServer),
				Value: gauge.Value,
			})
		}
	}
	attrsCommon, free := adapter[0].convert(gaugeGroup)
	tmpResults = make([]*AdaptedResult, 0, 2)
	if len(gaugesExceptRequestCount) > 0 {
		// for request count
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType:    Metric,
			AttrsList:     attrsCommon,
			Gauges:        gaugesExceptRequestCount,
			Timestamp:     gaugeGroup.Timestamp,
			FreeAttrsList: free,
		})
	}
	if len(requestCount) > 0 {
		attrsWithSlow, free := adapter[1].convert(gaugeGroup)
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType:    Metric,
			AttrsList:     attrsWithSlow,
			Gauges:        requestCount,
			Timestamp:     gaugeGroup.Timestamp,
			FreeAttrsList: free,
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
) *NetGaugeGroupAdapter {
	return &NetGaugeGroupAdapter{
		NetAdapterManager: createNetAdapterManager(customLabels),
		NetAdapterConfig:  config,
	}
}
