package aggregateprocessor

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal"
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal/defaultaggregator"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const Type = "aggregateprocessor"

type AggregateProcessor struct {
	cfg          *Config
	telemetry    *component.TelemetryTools
	nextConsumer consumer.Consumer

	aggregator  internal.Aggregator
	labelFilter *internal.LabelFilter
	stopCh      chan struct{}
	ticker      *time.Ticker
}

func New(config interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	cfg := config.(*Config)
	p := &AggregateProcessor{
		cfg:          cfg,
		telemetry:    telemetry,
		nextConsumer: nextConsumer,

		aggregator:  defaultaggregator.NewDefaultAggregator(cfg.AggregateKindMap),
		labelFilter: internal.NewLabelFilter(cfg.FilterLabels...),
		stopCh:      make(chan struct{}),
		ticker:      time.NewTicker(time.Duration(cfg.TickerInterval) * time.Second),
	}
	go p.runTicker()
	return p
}

// TODO: Graceful shutdown
func (p *AggregateProcessor) runTicker() {
	for {
		select {
		case <-p.stopCh:
			return
		case <-p.ticker.C:
			aggResults := p.aggregator.Dump()
			for _, agg := range aggResults {
				err := p.nextConsumer.Consume(agg)
				if err != nil {
					p.telemetry.Logger.Warn("Error happened when consuming aggregated recordersMap",
						zap.Error(err))
				}
			}
		}
	}
}

func (p *AggregateProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	switch gaugeGroup.Name {
	case constnames.NetRequestGaugeGroupName:
		var abnormalDataErr error
		// The abnormal recordersMap will be treated as trace in later processing.
		// Must trace be merged into metrics in this place? Yes, because we have to generate histogram metrics,
		// trace recordersMap should not be recorded again, otherwise the percentiles will be much higher.
		if p.isSampled(gaugeGroup) {
			gaugeGroup.Name = constnames.SingleNetRequestGaugeGroup
			abnormalDataErr = p.nextConsumer.Consume(gaugeGroup)
		}
		gaugeGroup.Name = constnames.AggregatedNetRequestGaugeGroup
		// Add a request_count metric
		gaugeGroup.Values = append(gaugeGroup.Values, &model.Gauge{
			Name:  constvalues.RequestCount,
			Value: 1,
		})
		p.aggregate(gaugeGroup)
		return abnormalDataErr
	default:
		p.aggregate(gaugeGroup)
		return nil
	}
}

func (p *AggregateProcessor) aggregate(gaugeGroup *model.GaugeGroup) {
	p.aggregator.Aggregate(gaugeGroup, p.labelFilter)
}

func (p *AggregateProcessor) isSampled(gaugeGroup *model.GaugeGroup) bool {
	randSeed := rand.Intn(100)
	if isAbnormal(gaugeGroup) {
		if (randSeed < p.cfg.SamplingRate.SlowData) && gaugeGroup.Labels.GetBoolValue(constlabels.IsSlow) {
			return true
		}
		if (randSeed < p.cfg.SamplingRate.ErrorData) && gaugeGroup.Labels.GetBoolValue(constlabels.IsError) {
			return true
		}
	} else {
		if randSeed < p.cfg.SamplingRate.NormalData {
			return true
		}
	}
	return false
}

// shouldAggregate returns true if the gaugeGroup is slow or has errors.
func isAbnormal(g *model.GaugeGroup) bool {
	return g.Labels.GetBoolValue(constlabels.IsSlow) || g.Labels.GetBoolValue(constlabels.IsError) ||
		g.Labels.GetIntValue(constlabels.ErrorType) > constlabels.NoError
}
