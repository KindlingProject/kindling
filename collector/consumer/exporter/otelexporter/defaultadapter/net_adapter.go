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
		return n.dealWithPreAggGaugeGroups(gaugeGroup)
	case constnames.SingleNetRequestGaugeGroup:
		return n.dealWithSingleGaugeGroup(gaugeGroup)
	default:
		return nil, nil
	}
}

func (n *NetGaugeGroupAdapter) dealWithSingleGaugeGroup(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	requestTotalTime, ok := gaugeGroup.GetGauge(constvalues.RequestTotalTime)
	if !ok {
		return nil, nil
	}
	results := make([]*AdaptedResult, 0, 2)
	if n.StoreTraceAsSpan && isSlowOrError(gaugeGroup) {
		if attrs, err := n.traceToSpanAdapter.convert(gaugeGroup); err != nil {
			return nil, err
		} else {
			results = append(results, &AdaptedResult{
				ResultType: Trace,
				Attrs:      attrs,
				Gauges:     []*model.Gauge{requestTotalTime},
				Timestamp:  gaugeGroup.Timestamp,
			})
		}
	}
	if n.StoreTraceAsMetric {
		labels, err := n.traceToSpanAdapter.transform(gaugeGroup)
		if err != nil {
			return results, err
		}
		results = append(results, &AdaptedResult{
			ResultType: Metric,
			Attrs:      nil,
			Gauges: []*model.Gauge{{
				Name:  constnames.TraceAsMetric,
				Value: requestTotalTime.Value,
			}},
			Labels:    labels,
			Timestamp: gaugeGroup.Timestamp,
		})
	}

	return results, nil
}

func (n *NetGaugeGroupAdapter) dealWithPreAggGaugeGroups(gaugeGroup *model.GaugeGroup) ([]*AdaptedResult, error) {
	results := make([]*AdaptedResult, 0, 4)
	isServer := gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)
	srcNamespace := gaugeGroup.Labels.GetStringValue(constlabels.SrcNamespace)
	if n.StoreExternalSrcIP && srcNamespace == constlabels.ExternalClusterNamespace && isServer {
		externalAdapterCache := n.detailTopologyAdapter
		if externalTopology, err := n.createNetMetricResults(gaugeGroup, externalAdapterCache); err != nil {
			return nil, err
		} else {
			results = append(results, externalTopology...)
		}
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
	if metrics, err := n.createNetMetricResults(gaugeGroup, metricAdapterCache); err != nil {
		return nil, err
	} else {
		results = append(results, metrics...)
	}
	return results, nil
}

func (n *NetGaugeGroupAdapter) createNetMetricResults(gaugeGroup *model.GaugeGroup, adapter [2]*LabelConverter) (tmpResults []*AdaptedResult, err error) {
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
	// TODO deal with error
	attrsCommon, _ := adapter[0].convert(gaugeGroup)

	tmpResults = make([]*AdaptedResult, 0, 2)
	if len(gaugesExceptRequestCount) > 0 {
		// for request count
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType: Metric,
			Attrs:      attrsCommon,
			Gauges:     gaugesExceptRequestCount,
			Timestamp:  gaugeGroup.Timestamp,
		})
	}
	if len(requestCount) > 0 {
		attrsWithSlow, _ := adapter[1].convert(gaugeGroup)
		tmpResults = append(tmpResults, &AdaptedResult{
			ResultType: Metric,
			Attrs:      attrsWithSlow,
			Gauges:     requestCount,
			Timestamp:  gaugeGroup.Timestamp,
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

func isSlowOrError(g *model.GaugeGroup) bool {
	return g.Labels.GetBoolValue(constlabels.IsSlow) || g.Labels.GetBoolValue(constlabels.IsError)
}
