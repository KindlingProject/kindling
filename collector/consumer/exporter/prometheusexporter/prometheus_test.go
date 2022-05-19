package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/spf13/viper"
	"testing"
	"time"
)

func InitPrometheusExporter(t *testing.T) exporter.Exporter {
	configPath := "testdata/kindling-collector-config.yml"
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error happened when reading config: %v", err)
	}
	config := &Config{}
	err = viper.UnmarshalKey("exporters.prometheusexporter", config)
	if err != nil {
		t.Fatalf("error happened when unmarshaling config: %v", err)
	}
	return NewExporter(config, component.NewDefaultTelemetryTools())
}

func TestConsumeAggNetGaugeGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitPrometheusExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makePreAggNetGaugeGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func TestConsumeSingleNetGaugeGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitPrometheusExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makeSingleGaugeGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func makeSingleGaugeGroup(i int) *model.GaugeGroup {
	gaugesGroup := &model.GaugeGroup{
		Name: constnames.SingleNetRequestGaugeGroup,
		Values: []*model.Gauge{
			model.NewIntGauge(constvalues.ResponseIo, 1234567891),
			model.NewIntGauge(constvalues.RequestTotalTime, int64(i)),
			model.NewIntGauge(constvalues.RequestIo, 4500),
			model.NewIntGauge(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")

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
			model.NewIntGauge(constvalues.ResponseIo, 1234567891),
			model.NewHistogramGauge(constvalues.RequestTotalTime, &model.Histogram{
				Sum:                int64(i),
				Count:              100,
				ExplicitBoundaries: []int64{1e6, 2e6, 5e6, 1e7, 2e7, 5e7, 1e8, 2e8, 5e8, 1e9, 2e9, 5e9},
				BucketCounts:       []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			}),
			model.NewIntGauge(constvalues.RequestTotalTime, int64(i)),
			model.NewIntGauge(constvalues.RequestIo, 4500),
			model.NewIntGauge(constvalues.RequestCount, 100),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: uint64(time.Now().UnixNano()),
	}
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	gaugesGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	gaugesGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	gaugesGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	gaugesGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	gaugesGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")

	gaugesGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	gaugesGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	gaugesGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	gaugesGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return gaugesGroup
}
