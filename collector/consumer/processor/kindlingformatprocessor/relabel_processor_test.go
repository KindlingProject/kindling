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
		t.Errorf("Consume Gauge failed %s \n", err)
	}
}

func makeGaugeGroup(latency int64) *model.GaugeGroup {
	labels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
		constlabels.DstNode:             model.NewStringValue("test-node"),
		constlabels.DstNamespace:        model.NewStringValue("test-namespace"),
		constlabels.DstWorkloadKind:     model.NewStringValue("deployment"),
		constlabels.DstWorkloadName:     model.NewStringValue("test-deploy"),
		constlabels.DstPod:              model.NewStringValue("test-pod"),
		constlabels.DstContainer:        model.NewStringValue("test-container"),
		constlabels.DstIp:               model.NewStringValue("10.0.0.1"),
		constlabels.DstPort:             model.NewIntValue(80),
		constlabels.ContentKey:          model.NewStringValue("/test"),
		constlabels.ResponseContent:     model.NewIntValue(201),
		constlabels.IsSlow:              model.NewBoolValue(true),
		constlabels.HttpUrl:             model.NewStringValue("/test"),
		constlabels.HttpRequestPayload:  model.NewStringValue("GET /test HTTP/1.1"),
		constlabels.HttpApmTraceId:      model.NewStringValue("asd1231"),
		constlabels.HttpMethod:          model.NewStringValue("GET"),
		constlabels.HttpResponsePayload: model.NewStringValue("200 HTTP/1.1 adads"),
		constlabels.HttpStatusCode:      model.NewIntValue(200),
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
