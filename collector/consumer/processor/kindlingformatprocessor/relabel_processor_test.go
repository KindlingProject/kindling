package kindlingformatprocessor

import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/spf13/viper"
	"testing"
)

type nopExporter struct {
}

func (n *nopExporter) Consume(gaugeGroup *model.GaugeGroup) error {
	// DoNothing
	fmt.Printf("%v", gaugeGroup)
	return nil
}

func creatProcessor(configFile string) (*processor.Processor, error) {
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = viper.UnmarshalKey("processors.kindlingformatprocessor", config)
	if err != nil {
		return nil, err
	}
	nop := nopExporter{}
	relabelProcessor := NewRelabelProcessor(config, component.NewDefaultTelemetryTools(), &nop)
	return &relabelProcessor, nil
}

func TestRelabelProcessor_Consume(t *testing.T) {
	processor, err := creatProcessor("testdata/span.yml")
	if err != nil {
		t.Errorf("Create Prcessor failed %s \n", err)
	}
	relabel := (*processor).(*RelabelProcessor)

	err = relabel.Consume(makeGaugeGroup(15))
	if err != nil {
		t.Errorf("Consume Gauges failed %s \n", err)
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
		constlabels.IsSlow:          model.NewBoolValue(true),
	})

	latencyGauge := &model.Gauge{
		Name:  "kindling_entity_request_duration_nanoseconds",
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
