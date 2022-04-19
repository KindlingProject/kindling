package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.uber.org/multierr"
)

const ProcessorName = "kindlingformatprocessor"

// RelabelProcessor generates new model.GaugeGroup according to the documentation.
type RelabelProcessor struct {
	cfg          *Config
	nextConsumer consumer.Consumer
	telemetry    *component.TelemetryTools
}

func NewRelabelProcessor(cfg interface{}, telemetry *component.TelemetryTools, nextConsumer consumer.Consumer) processor.Processor {
	processorCfg := cfg.(*Config)
	return &RelabelProcessor{
		cfg:          processorCfg,
		nextConsumer: nextConsumer,
		telemetry:    telemetry,
	}
}

func (r *RelabelProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	switch gaugeGroup.Name {
	case constnames.SingleNetRequestGaugeGroup:
		return r.consumeSingleGroup(gaugeGroup)
	case constnames.AggregatedNetRequestGaugeGroup:
		var requestCountErr error
		var err error
		requestCount, ok := gaugeGroup.GetGauge(constvalues.RequestCount)
		if ok {
			requestCountGaugeGroup := model.NewGaugeGroup(gaugeGroup.Name, gaugeGroup.Labels, gaugeGroup.Timestamp, requestCount)
			requestCountErr = r.consumeAggregatedGroup(requestCountGaugeGroup, true)
			gaugeGroup.RemoveGauge(constvalues.RequestCount)
		}
		err = r.consumeAggregatedGroup(gaugeGroup, false)
		return multierr.Combine(requestCountErr, err)
	default:
		return nil
	}
}

func (r *RelabelProcessor) consumeSingleGroup(gaugeGroup *model.GaugeGroup) error {
	var traceErr error = nil
	var spanErr error = nil

	// There could be normal data enter this method with sampling it.
	// But only abnormal data would be stored as metric considering
	// its high cardinality.
	if r.cfg.NeedTraceAsMetric && isSlowOrError(gaugeGroup) {
		// Trace as Metric
		trace := newGauges(gaugeGroup)
		traceErr = r.nextConsumer.Consume(trace.Process(r.cfg, TraceName, TopologyTraceInstanceInfo,
			TopologyTraceK8sInfo, SrcContainerInfo, DstContainerInfo, ServiceProtocolInfo, TraceStatusInfo))
	}
	if r.cfg.NeedTraceAsResourceSpan {
		// Trace As Span
		span := newGauges(gaugeGroup)
		spanErr = r.nextConsumer.Consume(span.Process(r.cfg, SpanName, traceSpanInstanceInfo,
			TopologyTraceK8sInfo, traceSpanContainerInfo, SpanProtocolInfo, traceSpanValuesToLabel))
	}

	return multierr.Combine(traceErr, spanErr)
}

func (r *RelabelProcessor) consumeAggregatedGroup(gaugeGroup *model.GaugeGroup, addIsSlowLabel bool) error {
	common := newGauges(gaugeGroup)
	if gaugeGroup.Labels.GetBoolValue(constlabels.IsServer) {
		// Do not emit detail protocol metric at this version
		//protocol := newGauges(gaugeGroup)
		//protocolErr := r.nextConsumer.Consume(protocol.Process(r.cfg, ProtocolDetailMetricName, ServiceInstanceInfo, ServiceK8sInfo, ProtocolDetailInfo))
		relabelFuns1 := []Relabel{MetricName, ServiceInstanceInfo, ServiceK8sInfo, ServiceProtocolInfo}
		if addIsSlowLabel {
			relabelFuns1 = append(relabelFuns1, AddIsSlowLabel)
		}
		metricErr := r.nextConsumer.Consume(common.Process(r.cfg, relabelFuns1...))
		var metricErr2 error
		if r.cfg.StoreExternalSrcIP {
			srcNamespace := gaugeGroup.Labels.GetStringValue(constlabels.SrcNamespace)
			if srcNamespace == constlabels.ExternalClusterNamespace {
				// Use data from server-side to generate a topology metric only when the namespace is EXTERNAL.
				externalGaugeGroup := newGauges(gaugeGroup)
				// Here we have to modify the field "IsServer" to generate the metric.
				externalGaugeGroup.Labels.AddBoolValue(constlabels.IsServer, false)
				relabelFuns2 := []Relabel{MetricName, TopologyInstanceInfo,
					TopologyK8sInfo, DstContainerInfo, TopologyProtocolInfo}
				if addIsSlowLabel {
					relabelFuns2 = append(relabelFuns2, AddIsSlowLabel)
				}
				metricErr2 = r.nextConsumer.Consume(externalGaugeGroup.Process(r.cfg, relabelFuns2...))
				// In case of using the original data later, we reset the field "IsServer".
				externalGaugeGroup.Labels.AddBoolValue(constlabels.IsServer, true)
			}
		}
		return multierr.Combine(metricErr, metricErr2)
	} else {
		relabelFuns3 := []Relabel{MetricName, TopologyInstanceInfo, TopologyK8sInfo,
			SrcContainerInfo, DstContainerInfo, TopologyProtocolInfo}
		if addIsSlowLabel {
			relabelFuns3 = append(relabelFuns3, AddIsSlowLabel)
		}
		metricErr := r.nextConsumer.Consume(common.Process(r.cfg, relabelFuns3...))
		return metricErr
	}
}

func isSlowOrError(g *model.GaugeGroup) bool {
	return g.Labels.GetBoolValue(constlabels.IsSlow) || g.Labels.GetBoolValue(constlabels.IsError)
}
