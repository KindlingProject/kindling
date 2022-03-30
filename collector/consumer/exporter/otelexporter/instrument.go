package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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
	adapters                 map[string]*Adapter
}

func newInstrumentFactory(meter metric.Meter, logger *zap.Logger, customLabels []attribute.KeyValue) *instrumentFactory {
	return &instrumentFactory{
		instruments:              sync.Map{},
		gaugeAsyncInstrumentInit: sync.Map{},
		meter:                    meter,
		gaugeChan:                newGaugeChannel(2000),
		customLabels:             customLabels,
		logger:                   logger,
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

//
//func (i *instrumentFactory) recordTraceAsMetric(metricName string, singleGauge model.GaugeGroup) {
//	i.gaugeChan.put(&singleGauge, i.logger)
//	if i.isGaugeAsyncInitialized(metricName) {
//		return
//	}
//
//	metric.Must(i.meter).NewInt64GaugeObserver(metricName, func(ctx context.Context, result metric.Int64ObserverResult) {
//		if adapter, ok := i.adapters[metricName]; ok {
//			channel := i.gaugeChan.getChannel(metricName)
//			for {
//				select {
//				case gaugeGroup := <-channel:
//					attrs, _ := adapter.adapter(gaugeGroup.Labels, gaugeGroup)
//					result.Observe(gaugeGroup.Values[0].Value, attrs...)
//				default:
//					return
//				}
//			}
//		} else {
//			newAdapter, _ := newAdapterBuilder(topologyMetricDicList,
//				[][]dictionary{topologyInstanceMetricDicList, topologyDetailMetricDicList}).
//				withExtraLabels(entityProtocol, updateProtocolKey).
//				withValueToLabels(traceStatus, getTraceStatusLabels).
//				withConstLabels(i.customLabels).
//				build()
//			adapterRWMutex.Lock()
//			i.adapters[metricName] = newAdapter
//			adapterRWMutex.Unlock()
//			channel := i.gaugeChan.getChannel(metricName)
//			for {
//				select {
//				case gaugeGroup := <-channel:
//					attrs, _ := newAdapter.adapter(gaugeGroup.Labels, gaugeGroup)
//					result.Observe(gaugeGroup.Values[0].Value, attrs...)
//				default:
//					return
//				}
//			}
//		}
//	}, WithDescription(metricName))
//}

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
