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
		var metricErr2 error
		if r.cfg.StoreExternalSrcIP {
			srcNamespace := gaugeGroup.Labels.GetStringValue(constlabels.SrcNamespace)
			if srcNamespace == constlabels.ExternalClusterNamespace {
				// Use data from server-side to generate a topology metric.
				// Here we have to modify the field "IsServer" to generate the metric.
				common.Labels.AddBoolValue(constlabels.IsServer, false)
				metricErr2 = r.nextConsumer.Consume(common.Process(r.cfg, MetricName, TopologyInstanceInfo, TopologyK8sInfo, TopologyProtocolInfo))
				// In case of using the original data later, we reset the field "IsServer".
				common.Labels.AddBoolValue(constlabels.IsServer, true)
			}
		}
		return multierr.Combine(traceErr, metricErr, metricErr2)
	} else {
		metricErr := r.nextConsumer.Consume(common.Process(r.cfg, MetricName, TopologyInstanceInfo, TopologyK8sInfo, SrcDockerInfo, TopologyProtocolInfo))
		return multierr.Combine(traceErr, metricErr)
	}
}
