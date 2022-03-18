package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

var gaugeGroupReceivedMetricName = "kindling_telemetry_metrics_gaugegroups_exporter_received_total"
var metricExportedCountMetricsName = "kindling_telemetry_metrics_metrics_exporter_exported_total"

var once sync.Once

var labelsSet map[labelKey]bool
var labelsSetMutex sync.RWMutex

var gaugeGroupReceiverCounter metric.Int64Counter
var metricExportedCountMetrics metric.Int64UpDownCounterObserver

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		gaugeGroupReceiverCounter = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(gaugeGroupReceivedMetricName)
		metricExportedCountMetrics = metric.Must(meterProvider.Meter("kindling")).NewInt64UpDownCounterObserver(
			metricExportedCountMetricsName, func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(len(labelsSet)))
			})
	})
}

func storeGaugeGroupKeys(group *model.GaugeGroup) {
	key := labelKey{
		metric:          "",
		srcIp:           group.Labels.GetStringValue(constlabels.SrcIp),
		dstIp:           group.Labels.GetStringValue(constlabels.DstIp),
		dstPort:         group.Labels.GetIntValue(constlabels.DstPort),
		requestContent:  group.Labels.GetStringValue(constlabels.ResponseContent),
		responseContent: group.Labels.GetStringValue(constlabels.ResponseContent),
		statusCode:      group.Labels.GetStringValue(constlabels.StatusCode),
		protocol:        group.Labels.GetStringValue(constlabels.Protocol),
	}
	labelsSetMutex.Lock()
	if _, ok := labelsSet[key]; ok {
		return
	}
	for _, value := range group.Values {
		key.metric = value.Name
		labelsSet[key] = true
	}
	labelsSetMutex.Unlock()
}
