package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/Kindling-project/kindling/collector/observability/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"testing"
	"time"
)

func TestAdapter_adapter(t *testing.T) {

	cfg := &Config{
		ExportKind:   PrometheusKindExporter,
		PromCfg:      &PrometheusConfig{Port: ":8080"},
		OtlpGrpcCfg:  nil,
		StdoutCfg:    nil,
		CustomLabels: nil,
		MetricAggregationMap: map[string]MetricAggregationKind{
			"kindling_entity_request_duration_nanoseconds":   2,
			"kindling_entity_request_send_bytes_total":       1,
			"kindling_entity_request_receive_bytes_total":    1,
			"kindling_topology_request_duration_nanoseconds": 2,
			"kindling_topology_request_request_bytes_total":  1,
			"kindling_topology_request_response_bytes_total": 1,
			"kindling_trace_request_duration_nanoseconds":    0,
			"kindling_tcp_rtt_milliseconds":                  0,
			"kindling_tcp_retransmit_total":                  1,
			"kindling_tcp_packet_loss_total":                 1,
		},
	}

	constLabels := []attribute.KeyValue{
		attribute.String("constLabels1", "constValues1"),
		attribute.Int("constLabels2", 2),
	}
	baseAdapterManager := createBaseAdapterManager(constLabels)
	detailTopologyAttrs, _ := baseAdapterManager.detailTopologyAdapter.adapter(makeOriginGaugeGroup(300000))
	//for i := 0; i < len(detailTopologyAttrs); i++ {
	//	fmt.Printf("%+v\n", detailTopologyAttrs[i])
	//}
	logger := logger.CreateFileRotationLogger(&lumberjack.Logger{
		Filename:   "test.log",
		MaxSize:    500,
		MaxAge:     10,
		MaxBackups: 1,
		LocalTime:  true,
		Compress:   false,
	})

	config := prometheus.Config{}
	// Create a meter

	c := controller.New(
		otelprocessor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
		),
	)
	exp, _ := prometheus.New(config, c)

	otelexporter := &OtelExporter{
		cfg:                  cfg,
		metricController:     c,
		traceProvider:        nil,
		defaultTracer:        nil,
		customLabels:         constLabels,
		instrumentFactory:    newInstrumentFactory(exp.MeterProvider().Meter(MeterName), logger, constLabels),
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            component.NewDefaultTelemetryTools(),
		adapterManager:       createBaseAdapterManager(constLabels),
	}

	go func() {
		err := StartServer(exp, otelexporter.telemetry.Logger, cfg.PromCfg.Port)
		if err != nil {
			otelexporter.telemetry.Logger.Warn("error starting otelexporter prometheus server: ", zap.Error(err))
		}
	}()

	for {
		otelexporter.instrumentFactory.meter.RecordBatch(context.Background(), detailTopologyAttrs, otelexporter.GetMetricMeasurement(makeOriginGaugeGroup(300000), false)...)
		time.Sleep(1 * time.Second)
	}
}

func TestNegativeValue_adapter(t *testing.T) {
	cfg := &Config{
		ExportKind:   StdoutKindExporter,
		PromCfg:      nil,
		OtlpGrpcCfg:  nil,
		StdoutCfg:    &StdoutConfig{CollectPeriod: 15 * time.Second},
		CustomLabels: nil,
		MetricAggregationMap: map[string]MetricAggregationKind{
			"kindling_entity_request_duration_nanoseconds":   2,
			"kindling_entity_request_send_bytes_total":       1,
			"kindling_entity_request_receive_bytes_total":    1,
			"kindling_topology_request_duration_nanoseconds": 2,
			"kindling_topology_request_request_bytes_total":  1,
			"kindling_topology_request_response_bytes_total": 1,
			"kindling_trace_request_duration_nanoseconds":    0,
			"kindling_tcp_rtt_milliseconds":                  0,
			"kindling_tcp_retransmit_total":                  1,
			"kindling_tcp_packet_loss_total":                 1,
		},
	}

	constLabels := []attribute.KeyValue{
		attribute.String("constLabels1", "constValues1"),
		attribute.Int("constLabels2", 2),
	}
	baseAdapterManager := createBaseAdapterManager(constLabels)
	logger := logger.CreateFileRotationLogger(&lumberjack.Logger{
		Filename:   "test.log",
		MaxSize:    500,
		MaxAge:     10,
		MaxBackups: 1,
		LocalTime:  true,
		Compress:   false,
	})
	exporter, _ := newExporters(context.Background(), cfg, logger)

	cont := controller.New(
		otelprocessor.NewFactory(simple.NewWithHistogramDistribution(
			histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
		), exporter.metricExporter),
		controller.WithExporter(exporter.metricExporter),
		controller.WithCollectPeriod(cfg.StdoutCfg.CollectPeriod),
		controller.WithResource(nil),
	)

	otelexporter := &OtelExporter{
		metricController:     cont,
		traceProvider:        nil,
		defaultTracer:        nil,
		customLabels:         nil,
		instrumentFactory:    newInstrumentFactory(cont.Meter(MeterName), logger, nil),
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            component.NewDefaultTelemetryTools(),
	}

	if err := cont.Start(context.Background()); err != nil {
		logger.Panic("failed to start controller:", zap.Error(err))
	}

	gauges := makeOriginGaugeGroup(-30)

	attrs, _ := baseAdapterManager.detailTopologyAdapter.adapter(gauges)
	otelexporter.instrumentFactory.meter.RecordBatch(context.Background(), attrs, otelexporter.GetMetricMeasurement(gauges, false)...)

}

func makeOriginGaugeGroup(latency int64) *model.GaugeGroup {
	labels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
		constlabels.DstNode:             model.NewStringValue("Dst-node"),
		constlabels.DstNamespace:        model.NewStringValue("Dst-namespace"),
		constlabels.DstWorkloadKind:     model.NewStringValue("Dst-deployment"),
		constlabels.DstWorkloadName:     model.NewStringValue("Dst-deploy"),
		constlabels.DstPod:              model.NewStringValue("Dst-pod"),
		constlabels.DstContainer:        model.NewStringValue("Dst-container"),
		constlabels.DstContainerId:      model.NewStringValue("Dst-containerid"),
		constlabels.DstIp:               model.NewStringValue("10.0.0.1"),
		constlabels.DstPort:             model.NewIntValue(80),
		constlabels.SrcNode:             model.NewStringValue("Src-node"),
		constlabels.SrcNamespace:        model.NewStringValue("Src-namespace"),
		constlabels.SrcWorkloadKind:     model.NewStringValue("Src-deployment"),
		constlabels.SrcWorkloadName:     model.NewStringValue("Src-deploy"),
		constlabels.SrcPod:              model.NewStringValue("Src-pod"),
		constlabels.SrcContainer:        model.NewStringValue("Src-container"),
		constlabels.SrcContainerId:      model.NewStringValue("Src-containerid"),
		constlabels.SrcIp:               model.NewStringValue("10.0.0.2"),
		constlabels.SrcPort:             model.NewIntValue(36002),
		constlabels.ContentKey:          model.NewStringValue("/test"),
		constlabels.ResponseContent:     model.NewIntValue(201),
		constlabels.IsSlow:              model.NewBoolValue(true),
		constlabels.HttpUrl:             model.NewStringValue("/test"),
		constlabels.HttpRequestPayload:  model.NewStringValue("GET /test HTTP/1.1"),
		constlabels.HttpApmTraceId:      model.NewStringValue("asd1231"),
		constlabels.HttpMethod:          model.NewStringValue("GET"),
		constlabels.HttpResponsePayload: model.NewStringValue("200 HTTP/1.1 adads"),
		constlabels.HttpStatusCode:      model.NewIntValue(200),
		constlabels.Protocol:            model.NewStringValue(constvalues.ProtocolHttp),
	})

	latencyGauge := &model.Gauge{
		Name:  constvalues.RequestTotalTime,
		Value: latency,
	}

	requestTimeGauge := &model.Gauge{
		Name:  constvalues.RequestSentTime,
		Value: 300,
	}
	waitTtfbTime := &model.Gauge{
		Name:  constvalues.WaitingTtfbTime,
		Value: 400,
	}
	contentDownload := &model.Gauge{
		Name:  constvalues.ContentDownloadTime,
		Value: 500,
	}
	connectTime := &model.Gauge{
		Name:  constvalues.ConnectTime,
		Value: 600,
	}
	reqio := &model.Gauge{
		Name:  constvalues.RequestIo,
		Value: 700,
	}

	gaugeGroup := model.NewGaugeGroup("", labels, 0, latencyGauge, requestTimeGauge, waitTtfbTime, contentDownload, connectTime, reqio)
	return gaugeGroup
}
