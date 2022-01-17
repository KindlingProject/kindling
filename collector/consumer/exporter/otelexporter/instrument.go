package otelexporter

import (
	"context"
	"github.com/dxsup/kindling-collector/model"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"sync"

	"go.opentelemetry.io/otel/metric"
)

type instrumentFactory struct {
	// TODO: Initialize instruments when initializing factory
	instruments              sync.Map
	gaugeAsyncInstrumentInit sync.Map
	meter                    metric.Meter
	gaugeChan                *gaugeChannel
	customLabels             []attribute.KeyValue
	logger                   *zap.Logger
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
	})
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
