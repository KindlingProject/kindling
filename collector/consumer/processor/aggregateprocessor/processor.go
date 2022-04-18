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
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const Type = "aggregateprocessor"

type AggregateProcessor struct {
	cfg          *Config
	telemetry    *component.TelemetryTools
	nextConsumer consumer.Consumer

	aggregator               internal.Aggregator
	netRequestLabelSelectors *internal.LabelSelectors
	tcpLabelSelectors        *internal.LabelSelectors
	stopCh                   chan struct{}
	ticker                   *time.Ticker
}

func New(config interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	cfg := config.(*Config)
	p := &AggregateProcessor{
		cfg:          cfg,
		telemetry:    telemetry,
		nextConsumer: nextConsumer,

		aggregator:               defaultaggregator.NewDefaultAggregator(toAggregatedConfig(cfg.AggregateKindMap)),
		netRequestLabelSelectors: newNetRequestLabelSelectors(),
		tcpLabelSelectors:        newTcpLabelSelectors(),
		stopCh:                   make(chan struct{}),
		ticker:                   time.NewTicker(time.Duration(cfg.TickerInterval) * time.Second),
	}
	go p.runTicker()
	return p
}

func toAggregatedConfig(m map[string][]AggregatedKindConfig) *defaultaggregator.AggregatedConfig {
	ret := &defaultaggregator.AggregatedConfig{KindMap: make(map[string][]defaultaggregator.KindConfig)}
	for k, v := range m {
		kindConfig := make([]defaultaggregator.KindConfig, len(v))
		for i, kind := range v {
			if kind.OutputName == "" {
				kind.OutputName = k
			}
			kindConfig[i] = defaultaggregator.KindConfig{
				OutputName: kind.OutputName,
				Kind:       defaultaggregator.GetAggregatorKind(kind.Kind),
			}
		}
		ret.KindMap[k] = kindConfig
	}
	return ret
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
		p.aggregator.Aggregate(gaugeGroup, p.netRequestLabelSelectors)
		return abnormalDataErr
	case constnames.TcpGaugeGroupName:
		p.aggregator.Aggregate(gaugeGroup, p.tcpLabelSelectors)
		return nil
	default:
		p.aggregator.Aggregate(gaugeGroup, p.netRequestLabelSelectors)
		return nil
	}
}

// TODO: make it configurable instead of hard-coded
func newNetRequestLabelSelectors() *internal.LabelSelectors {
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

		internal.LabelSelector{Name: constlabels.IsSlow, VType: internal.BooleanType},
		internal.LabelSelector{Name: constlabels.HttpStatusCode, VType: internal.IntType},
		internal.LabelSelector{Name: constlabels.DnsRcode, VType: internal.IntType},
		internal.LabelSelector{Name: constlabels.SqlErrCode, VType: internal.IntType},
		internal.LabelSelector{Name: constlabels.ContentKey, VType: internal.StringType},
		internal.LabelSelector{Name: constlabels.DnsDomain, VType: internal.StringType},
		internal.LabelSelector{Name: constlabels.KafkaTopic, VType: internal.StringType},
	)
}

func newTcpLabelSelectors() *internal.LabelSelectors {
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
