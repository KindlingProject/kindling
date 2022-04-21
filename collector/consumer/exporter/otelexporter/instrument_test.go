package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/Kindling-project/kindling/collector/observability/logger"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"gopkg.in/natefinch/lumberjack.v2"
	"testing"
	"time"
)

func Test_instrumentFactory_recordLastValue(t *testing.T) {
	cfg := &Config{
		ExportKind:   StdoutKindExporter,
		PromCfg:      nil,
		OtlpGrpcCfg:  nil,
		StdoutCfg:    &StdoutConfig{CollectPeriod: 10 * time.Second},
		CustomLabels: nil,
		MetricAggregationMap: map[string]MetricAggregationKind{
			"kindling_entity_request_duration_nanoseconds":         2,
			"kindling_entity_request_send_bytes_total":             1,
			"kindling_entity_request_receive_bytes_total":          1,
			"kindling_topology_request_duration_nanoseconds_total": 2,
			"kindling_topology_request_request_bytes_total":        1,
			"kindling_topology_request_response_bytes_total":       1,
			"kindling_trace_request_duration_nanoseconds":          0,
			"kindling_tcp_rtt_microseconds":                        0,
			"kindling_tcp_retransmit_total":                        1,
			"kindling_tcp_packet_loss_total":                       1,
		},
		AdapterConfig: &AdapterConfig{
			NeedTraceAsResourceSpan: true,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      true,
		},
	}

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

	cont.Start(context.Background())

	ins := newInstrumentFactory(cont.Meter("test"), component.NewDefaultTelemetryTools().Logger, nil)

	go func() {
		for i := 0; i < 10000; i++ {
			time.Sleep(1 * time.Second)
			group := makeTcpGroup(int64(i))
			ins.recordLastValue("kindling_tcp_rtt_microseconds", group)
		}
	}()

	for i := 0; i < 10000; i++ {
		time.Sleep(1 * time.Second)
		group := makeTraceAsMetric(int64(i))
		attrs, _ := defaultadapter.NewNetAdapter(nil, &defaultadapter.NetAdapterConfig{
			StoreTraceAsMetric: true,
			StoreTraceAsSpan:   false,
			StorePodDetail:     true,
			StoreExternalSrcIP: true,
		}).Transform(group)
		ins.recordLastValue(constnames.TraceAsMetric, model.NewGaugeGroup(PreAggMetric, attrs, group.Timestamp, group.Values...))
	}
}

func makeTcpGroup(rttLatency int64) *model.GaugeGroup {
	return model.NewGaugeGroup(
		constnames.TcpGaugeGroupName,
		model.NewAttributeMapWithValues(
			map[string]model.AttributeValue{
				// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.SrcPort: model.NewIntValue(33333),
				constlabels.DstIp:   model.NewStringValue("dst-ip"),
				constlabels.DstPort: model.NewIntValue(8080),
				// k8sInfo
				constlabels.DstPod:          model.NewStringValue("dst-pod"),
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),

				// isSlow
				constlabels.IsSlow: model.NewBoolValue(false),
			}),
		123,
		[]*model.Gauge{
			{constvalues.TcpRttMetricName, rttLatency},
		}...)
}

func makeNetGroup(requestLatency int64) *model.GaugeGroup {
	return model.NewGaugeGroup(
		constnames.AggregatedNetRequestGaugeGroup,
		model.NewAttributeMapWithValues(
			map[string]model.AttributeValue{
				// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.SrcPort: model.NewIntValue(33333),
				constlabels.DstIp:   model.NewStringValue("dst-ip"),
				constlabels.DstPort: model.NewIntValue(8080),

				// protocolInfo
				constlabels.Protocol:       model.NewStringValue("http"),
				constlabels.HttpUrl:        model.NewStringValue("/test"),
				constlabels.HttpStatusCode: model.NewIntValue(200),

				// k8sInfo
				constlabels.DstPod:          model.NewStringValue("dst-pod"),
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),

				// isSlow
				constlabels.IsSlow: model.NewBoolValue(false),
			}),
		123,
		[]*model.Gauge{
			{constvalues.RequestTotalTime, requestLatency},
		}...)
}

func makeTraceAsMetric(requestLatency int64) *model.GaugeGroup {
	return model.NewGaugeGroup(
		constnames.AggregatedNetRequestGaugeGroup,
		model.NewAttributeMapWithValues(
			map[string]model.AttributeValue{
				// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.SrcPort: model.NewIntValue(33333),
				constlabels.DstIp:   model.NewStringValue("dst-ip"),
				constlabels.DstPort: model.NewIntValue(8080),

				// protocolInfo
				constlabels.Protocol:       model.NewStringValue("http"),
				constlabels.HttpUrl:        model.NewStringValue("/test"),
				constlabels.HttpStatusCode: model.NewIntValue(200),

				// k8sInfo
				constlabels.DstPod:          model.NewStringValue("dst-pod"),
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),

				// isSlow
				constlabels.IsSlow: model.NewBoolValue(false),
			}),
		123,
		[]*model.Gauge{
			{constnames.TraceAsMetric, requestLatency},
		}...)
}
