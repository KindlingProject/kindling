package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/internal"
	"github.com/Kindling-project/kindling/collector/internal/defaultaggregator"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"sync"

	"go.opentelemetry.io/otel/metric"
)

var traceAsMetricHelp = "Describe a single request which is abnormal. " +
	"For status labels, number '1', '2' and '3' stands for Green, Yellow and Red respectively."

var adapterRWMutex sync.RWMutex

type instrumentFactory struct {
	// TODO: Initialize instruments when initializing factory
	instruments              sync.Map
	gaugeAsyncInstrumentInit sync.Map
	meter                    metric.Meter
	gaugeChan                *gaugeChannel
	customLabels             []attribute.KeyValue
	logger                   *zap.Logger

	aggregators            sync.Map
	adapterManager         *BaseAdapterManager
	traceAsMetricAggConfig *defaultaggregator.AggregatedConfig
}

func newInstrumentFactory(meter metric.Meter, logger *zap.Logger, customLabels []attribute.KeyValue) *instrumentFactory {
	return &instrumentFactory{
		instruments:              sync.Map{},
		gaugeAsyncInstrumentInit: sync.Map{},
		meter:                    meter,
		gaugeChan:                newGaugeChannel(2000),
		customLabels:             customLabels,
		logger:                   logger,
		adapterManager:           createBaseAdapterManager(customLabels),
		traceAsMetricAggConfig: &defaultaggregator.AggregatedConfig{
			KindMap: map[string][]defaultaggregator.KindConfig{
				constvalues.RequestTotalTime: []defaultaggregator.KindConfig{
					{Kind: defaultaggregator.LastKind, OutputName: constlabels.ToKindlingTraceAsMetricName()},
				},
			}},
	}
}
func (i *instrumentFactory) getInstrument(metricName string, kind MetricAggregationKind) instrument {
	if ins, ok := i.instruments.Load(metricName); ok {
		return ins.(instrument)
	} else {
		newIns := i.createNewInstrument(metricName, kind)
		i.instruments.Store(metricName, newIns)
		return newIns
	}
}

func (i *instrumentFactory) createNewInstrument(metricName string, kind MetricAggregationKind) instrument {
	switch kind {
	case MACounterKind:
		ins := metric.Must(i.meter).NewInt64Counter(metricName)
		return &counterInstrument{instrument: &ins}
	case MAHistogramKind:
		ins := metric.Must(i.meter).NewInt64Histogram(metricName)
		return &histogramInstrument{instrument: &ins}
	default:
		ins := metric.Must(i.meter).NewInt64Counter(metricName)
		return &counterInstrument{instrument: &ins}
	}
}

// only support TraceAsMetric and TcpRttMills now
func (i *instrumentFactory) recordLastValue(metricName string, singleGauge *model.GaugeGroup) {
	if item, ok := i.aggregators.Load(metricName); ok {
		aggregator := item.(*defaultaggregator.DefaultAggregator)
		aggregator.Aggregate(singleGauge, getSelector(metricName))
		return
	}

	newAggregator := defaultaggregator.NewDefaultAggregator(getAggConfig(metricName))
	i.aggregators.Store(metricName, newAggregator)
	newAggregator.Aggregate(singleGauge, getSelector(metricName))
	metric.Must(i.meter).NewInt64GaugeObserver(metricName, func(ctx context.Context, result metric.Int64ObserverResult) {
		var dumps []*model.GaugeGroup
		if item, ok := i.aggregators.Load(metricName); ok {
			aggregator := item.(*defaultaggregator.DefaultAggregator)
			dumps = aggregator.Dump()
			for s := 0; s < len(dumps); s++ {
				result.Observe(dumps[s].Values[0].Value, GetLabels(dumps[s].Labels, i.customLabels)...)
			}
		}
	})
}

func getAggConfig(metricName string) *defaultaggregator.AggregatedConfig {
	return &defaultaggregator.AggregatedConfig{
		KindMap: map[string][]defaultaggregator.KindConfig{
			metricName: {
				{Kind: defaultaggregator.LastKind, OutputName: metricName},
			},
		}}
}

func getSelector(metricName string) *internal.LabelSelectors {
	switch metricName {
	case constlabels.ToKindlingTraceAsMetricName():
		return internal.NewLabelSelectors(
			internal.LabelSelector{Name: constlabels.Pid, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.Protocol, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.IsServer, VType: internal.BooleanType},
			internal.LabelSelector{Name: constlabels.ContainerId, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcNode, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcNodeIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcNamespace, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcPod, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcWorkloadName, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcService, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcContainerId, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcContainer, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNode, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNodeIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNamespace, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstPod, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstWorkloadName, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstWorkloadKind, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstService, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstPort, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.DnatIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DnatPort, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.DstContainerId, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstContainer, VType: internal.StringType},

			internal.LabelSelector{Name: constlabels.RequestReqxferStatus, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.RequestProcessingStatus, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.ResponseRspxferStatus, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.RequestDurationStatus, VType: internal.StringType},

			internal.LabelSelector{Name: constlabels.RequestContent, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.ResponseContent, VType: internal.StringType},
		)
	case "kindling_tcp_rtt_microseconds":
		return internal.NewLabelSelectors(
			internal.LabelSelector{Name: constlabels.SrcNode, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcNodeIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcNamespace, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcPod, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcWorkloadName, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcService, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcPort, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.SrcContainerId, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.SrcContainer, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNode, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNodeIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstNamespace, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstPod, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstWorkloadName, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstWorkloadKind, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstService, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstPort, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.DnatIp, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DnatPort, VType: internal.IntType},
			internal.LabelSelector{Name: constlabels.DstContainerId, VType: internal.StringType},
			internal.LabelSelector{Name: constlabels.DstContainer, VType: internal.StringType},
		)

	}

	return nil
}

func (i *instrumentFactory) recordGaugeAsync(metricName string, singleGauge model.GaugeGroup) {
	i.gaugeChan.put(&singleGauge, i.logger)
	if i.isGaugeAsyncInitialized(metricName) {
		return
	}

	metric.Must(i.meter).NewInt64GaugeObserver(metricName, func(ctx context.Context, result metric.Int64ObserverResult) {
		channel := i.gaugeChan.getChannel(metricName)
		for {
			select {
			case gaugeGroup := <-channel:
				labels := gaugeGroup.Labels
				result.Observe(gaugeGroup.Values[0].Value, GetLabels(labels, i.customLabels)...)
			default:
				return
			}
		}
	}, WithDescription(metricName))
}

func WithDescription(metricName string) metric.InstrumentOption {
	var option metric.InstrumentOption
	switch metricName {
	case constlabels.ToKindlingTraceAsMetricName():
		option = metric.WithDescription(traceAsMetricHelp)
	default:
		option = metric.WithDescription("")
	}
	return option
}

func (i *instrumentFactory) isGaugeAsyncInitialized(metricName string) bool {
	_, ok := i.gaugeAsyncInstrumentInit.Load(metricName)
	if ok {
		return true
	} else {
		i.gaugeAsyncInstrumentInit.Store(metricName, true)
		return false
	}
}

type instrument interface {
	Measurement(value int64) metric.Measurement
}

type counterInstrument struct {
	instrument *metric.Int64Counter
}

func (c *counterInstrument) Measurement(value int64) metric.Measurement {
	return c.instrument.Measurement(value)
}

type histogramInstrument struct {
	instrument *metric.Int64Histogram
}

func (h *histogramInstrument) Measurement(value int64) metric.Measurement {
	return h.instrument.Measurement(value)
}

type AsyncGaugeGroup struct {
}
