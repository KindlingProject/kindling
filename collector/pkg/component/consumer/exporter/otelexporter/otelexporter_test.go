package otelexporter

import (
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"
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

func TestConsumeAggNetMetricGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitOtelExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makePreAggNetMetricGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func TestConsumeSingleNetMetricGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitOtelExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makeSingleMetricGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func makeSingleMetricGroup(i int) *model.DataGroup {
	metricsGroup := &model.DataGroup{
		Name: constnames.SingleNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.ResponseIo, 1234567891),
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return metricsGroup
}

func makePreAggNetMetricGroup(i int) *model.DataGroup {
	metricsGroup := &model.DataGroup{
		Name: constnames.AggregatedNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.ResponseIo, 1234567891),
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return metricsGroup
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
			"kindling_entity_request_total":                          MACounterKind,
			"kindling_entity_request_duration_nanoseconds_total":     MACounterKind,
			"kindling_entity_request_average_duration_nanoseconds":   MAHistogramKind,
			"kindling_entity_request_send_bytes_total":               MACounterKind,
			"kindling_entity_request_receive_bytes_total":            MACounterKind,
			"kindling_topology_request_total":                        MACounterKind,
			"kindling_topology_request_duration_nanoseconds_total":   MACounterKind,
			"kindling_topology_request_average_duration_nanoseconds": MAHistogramKind,
			"kindling_topology_request_request_bytes_total":          MACounterKind,
			"kindling_topology_request_response_bytes_total":         MACounterKind,
			"kindling_trace_request_duration_nanoseconds":            MAGaugeKind,
			"kindling_tcp_srtt_microseconds":                         MAGaugeKind,
			"kindling_tcp_retransmit_total":                          MACounterKind,
			"kindling_tcp_packet_loss_total":                         MACounterKind,
		},
		AdapterConfig: &AdapterConfig{
			NeedTraceAsResourceSpan: true,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      true,
		},
	}

	telemetry := component.NewDefaultTelemetryTools()
	exporter, _ := newExporters(context.Background(), cfg, telemetry)

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
		instrumentFactory:    newInstrumentFactory(cont.Meter(MeterName), telemetry, nil),
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            component.NewDefaultTelemetryTools(),
		adapters: []adapter.Adapter{
			adapter.NewNetAdapter(nil, &adapter.NetAdapterConfig{
				StoreTraceAsMetric: cfg.AdapterConfig.NeedTraceAsMetric,
				StoreTraceAsSpan:   cfg.AdapterConfig.NeedTraceAsResourceSpan,
				StorePodDetail:     cfg.AdapterConfig.NeedPodDetail,
				StoreExternalSrcIP: cfg.AdapterConfig.StoreExternalSrcIP,
			}),
			adapter.NewSimpleAdapter([]string{constnames.TcpMetricGroupName}, nil),
		},
	}

	if err := cont.Start(context.Background()); err != nil {
		telemetry.Logger.Panic("failed to start controller:", zap.Error(err))
	}
	newSelfMetrics(otelexporter.telemetry.MeterProvider)

	//mockMetric := make(chan *model.DataGroup, 100)
	//MockMetric(mockMetric, 800, 1000, 10*time.Minute)
	recordCounter := 0

	metricsGroupsSlice := make([]*model.DataGroup, dimension)

	for i := 0; i < dimension; i++ {
		metricsGroup := &model.DataGroup{
			Name: constnames.AggregatedNetRequestMetricGroup,
			Metrics: []*model.Metric{
				model.NewIntMetric(constvalues.ResponseIo, 1234567891),
				model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
				model.NewIntMetric(constvalues.RequestIo, 4500),
				model.NewIntMetric(constvalues.RequestCount, 4500),
			},
			Labels:    model.NewAttributeMap(),
			Timestamp: 19900909090,
		}
		metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

		metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

		metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
		metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

		// Topology data preferentially use D Nat Ip and D Nat Port
		metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
		metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)

		metricsGroupsSlice[i] = metricsGroup
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		otelexporter.Consume(metricsGroupsSlice[recordCounter%dimension])
		recordCounter++
	}

	log.Printf("Test Finished!")
}
