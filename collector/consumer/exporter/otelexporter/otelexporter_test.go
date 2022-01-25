package otelexporter

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/spf13/viper"
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

func TestConsumeGaugeGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitOtelExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makeGaugeGroup(latency))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func makeGaugeGroup(latency int64) *model.GaugeGroup {
	labels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
		constlabels.Node:            model.NewStringValue("test-node"),
		constlabels.Namespace:       model.NewStringValue("test-namespace"),
		constlabels.WorkloadKind:    model.NewStringValue("deployment"),
		constlabels.WorkloadName:    model.NewStringValue("test-deploy"),
		constlabels.Pod:             model.NewStringValue("test-pod"),
		constlabels.Container:       model.NewStringValue("test-container"),
		constlabels.Ip:              model.NewStringValue("10.0.0.1"),
		constlabels.Port:            model.NewIntValue(80),
		constlabels.RequestContent:  model.NewStringValue("/test"),
		constlabels.ResponseContent: model.NewIntValue(201),
	})

	latencyGauge := model.Gauge{
		Name:  "kindling_entity_request_duration_nanoseconds",
		Value: latency,
	}

	gaugeGroup := model.NewGaugeGroup("", labels, 0, latencyGauge)
	return gaugeGroup
}
