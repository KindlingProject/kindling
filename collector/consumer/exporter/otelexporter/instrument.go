package otelexporter

import (
	"context"
	"fmt"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	defaultaggregator "github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
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

	aggregators    sync.Map
	adapterManager *BaseAdapterManager

	traceAsMetricSelector *aggregator.LabelSelectors
	TcpRttMillsSelector   *aggregator.LabelSelectors
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

		traceAsMetricSelector: newTraceAsMetricSelectors(),
		TcpRttMillsSelector:   newTcpRttMicroSecondsSelectors(),
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
		aggregator.Aggregate(singleGauge, i.getSelector(metricName))
		return
	}

	newAggregator := defaultaggregator.NewDefaultAggregator(getAggConfig(metricName))
	i.aggregators.Store(metricName, newAggregator)
	newAggregator.Aggregate(singleGauge, i.getSelector(metricName))
	metric.Must(i.meter).NewInt64GaugeObserver(metricName, func(ctx context.Context, result metric.Int64ObserverResult) {
		var dumps []*model.GaugeGroup
		if item, ok := i.aggregators.Load(metricName); ok {
			aggregator := item.(*defaultaggregator.DefaultAggregator)
			dumps = aggregator.Dump()
			for s := 0; s < len(dumps); s++ {
				if len(dumps[s].Values) > 0 {
					result.Observe(dumps[s].Values[0].Value, GetLabels(dumps[s].Labels, i.customLabels)...)
				} else {
					fmt.Println("Warning")
				}
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

func (i *instrumentFactory) getSelector(metricName string) *aggregator.LabelSelectors {
	switch metricName {
	case constlabels.ToKindlingTraceAsMetricName():
		return i.traceAsMetricSelector
	case constnames.TcpRttMetricName:
		return i.TcpRttMillsSelector
	default:
		return nil
	}
}

func newTraceAsMetricSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.Pid, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.Protocol, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.IsServer, VType: aggregator.BooleanType},
		aggregator.LabelSelector{Name: constlabels.ContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},

		aggregator.LabelSelector{Name: constlabels.RequestReqxferStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.RequestProcessingStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.ResponseRspxferStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.RequestDurationStatus, VType: aggregator.StringType},

		aggregator.LabelSelector{Name: constlabels.RequestContent, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.ResponseContent, VType: aggregator.StringType},
	)
}

func newTcpRttMicroSecondsSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},
	)
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
