package otelexporter

import (
	"context"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/Kindling-project/kindling/collector/observability/logger"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"strconv"
	"testing"
	"time"
)

func InitOtelExporter(t *testing.T) exporter.Exporter {
	configPath := "testdata/kindling-collector-config.yml"
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error happened when reading config: %v", err)
	}
	config := &Config{}
	err = viper.UnmarshalKey("exporters.otelexporter", config)
	if err != nil {
		t.Fatalf("error happened when unmarshaling config: %v", err)
	}
	return NewExporter(config, component.NewDefaultTelemetryTools())
}

func TestConsumeAggNetGaugeGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitOtelExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makePreAggNetGaugeGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func makeSingleGaugeGroup(i int) *model.GaugeGroup {
	gaugesGroup := &model.GaugeGroup{
		Name: constnames.SingleNetRequestGaugeGroup,
		Values: []*model.Gauge{
			{
				constvalues.ResponseIo,
				1234567891,
			},
			{
				constvalues.RequestTotalTime,
				int64(i),
			},
			{
				constvalues.RequestIo,
				4500,
			},
			{
				constvalues.RequestCount,
				4500,
			},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	gaugesGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return gaugesGroup
}

func makePreAggNetGaugeGroup(i int) *model.GaugeGroup {
	gaugesGroup := &model.GaugeGroup{
		Name: constnames.AggregatedNetRequestGaugeGroup,
		Values: []*model.Gauge{
			{
				constvalues.ResponseIo,
				1234567891,
			},
			{
				constvalues.RequestTotalTime,
				int64(i),
			},
			{
				constvalues.RequestIo,
				4500,
			},
			{
				constvalues.RequestCount,
				4500,
			},
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	gaugesGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return gaugesGroup
}

func BenchmarkOtelExporter_Consume(b *testing.B) {
	dimension := 100000

	cfg := &Config{
		ExportKind:   StdoutKindExporter,
		PromCfg:      nil,
		OtlpGrpcCfg:  nil,
		StdoutCfg:    &StdoutConfig{CollectPeriod: 30 * time.Second},
		CustomLabels: nil,
		MetricAggregationMap: map[string]MetricAggregationKind{
			"kindling_entity_request_duration_nanoseconds":         2,
			"kindling_entity_request_send_bytes_total":             1,
			"kindling_entity_request_receive_bytes_total":          1,
			"kindling_topology_request_duration_nanoseconds_total": 2,
			"kindling_topology_request_request_bytes_total":        1,
			"kindling_topology_request_response_bytes_total":       1,
			"kindling_trace_request_duration_nanoseconds":          0,
			"kindling_tcp_rtt_milliseconds":                        0,
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

	otelexporter := &OtelExporter{
		cfg:                  cfg,
		metricController:     cont,
		traceProvider:        nil,
		defaultTracer:        nil,
		customLabels:         nil,
		instrumentFactory:    newInstrumentFactory(cont.Meter(MeterName), logger, nil),
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            component.NewDefaultTelemetryTools(),
		adapters: []defaultadapter.Adapter{
			defaultadapter.NewNetAdapter(nil, &defaultadapter.NetAdapterConfig{
				StoreTraceAsMetric: cfg.AdapterConfig.NeedTraceAsMetric,
				StoreTraceAsSpan:   cfg.AdapterConfig.NeedTraceAsResourceSpan,
				StorePodDetail:     cfg.AdapterConfig.NeedPodDetail,
				StoreExternalSrcIP: cfg.AdapterConfig.StoreExternalSrcIP,
			}),
			defaultadapter.NewTcpAdapter(nil),
		},
	}

	if err := cont.Start(context.Background()); err != nil {
		logger.Panic("failed to start controller:", zap.Error(err))
	}
	newSelfMetrics(otelexporter.telemetry.MeterProvider)

	//mockMetric := make(chan *model.GaugeGroup, 100)
	//MockMetric(mockMetric, 800, 1000, 10*time.Minute)
	recordCounter := 0

	gaugesGroupsSlice := make([]*model.GaugeGroup, dimension)

	for i := 0; i < dimension; i++ {
		gaugesGroup := &model.GaugeGroup{
			Name: constnames.AggregatedNetRequestGaugeGroup,
			Values: []*model.Gauge{
				{
					constvalues.ResponseIo,
					1234567891,
				},
				{
					constvalues.RequestTotalTime,
					3300,
				},
				{
					constvalues.RequestIo,
					4500,
				},
				{
					constvalues.RequestCount,
					4500,
				},
			},
			Labels:    model.NewAttributeMap(),
			Timestamp: 19900909090,
		}
		gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

		gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
		gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

		gaugesGroup.Labels.AddStringValue(constlabels.Protocol, "http")
		gaugesGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

		// Topology data preferentially use D Nat Ip and D Nat Port
		gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
		gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)

		gaugesGroupsSlice[i] = gaugesGroup
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		otelexporter.Consume(gaugesGroupsSlice[recordCounter%dimension])
		recordCounter++
	}

	log.Printf("Test Finished!")
}
