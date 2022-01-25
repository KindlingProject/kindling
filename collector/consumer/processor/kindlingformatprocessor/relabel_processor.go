package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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
	return &RelabelProcessor{
		cfg:          cfg.(*Config),
		nextConsumer: nextConsumer,
		telemetry:    telemetry,
	}
}

func (r *RelabelProcessor) Consume(gaugeGroup *model.GaugeGroup) error {
	common := newGauges(gaugeGroup)

	var traceErr error = nil

	if r.cfg.NeedTraceAsMetric && common.isSlowOrError() {
		// Trace
		trace := newGauges(gaugeGroup)
		traceErr = r.nextConsumer.Consume(trace.Process(r.cfg, TraceName, TopologyTraceInstanceInfo, TopologyTraceK8sInfo, ServiceProtocolInfo, TraceStatusInfo))
	}

	// The data when the field is Error is true and the error Type is 2, do not generate metric
	errorType := gaugeGroup.Labels.GetIntValue(constlabels.ErrorType)
	if errorType == constlabels.ConnectFail || errorType == constlabels.NoResponse {
		return traceErr
	}

	if gaugeGroup.Labels.GetBoolValue(constlabels.IsServer) {
		// Do not emit detail protocol metric at this version
		//protocol := newGauges(gaugeGroup)
		//protocolErr := r.nextConsumer.Consume(protocol.Process(r.cfg, ProtocolDetailMetricName, ServiceInstanceInfo, ServiceK8sInfo, ProtocolDetailInfo))
		metricErr := r.nextConsumer.Consume(common.Process(r.cfg, MetricName, ServiceInstanceInfo, ServiceK8sInfo, ServiceProtocolInfo))
		return multierr.Combine(traceErr, metricErr)
	} else {
		metricErr := r.nextConsumer.Consume(common.Process(r.cfg, MetricName, TopologyInstanceInfo, TopologyK8sInfo, SrcDockerInfo, TopologyProtocolInfo))
		return multierr.Combine(traceErr, metricErr)
	}
}
