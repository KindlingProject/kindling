package otelexporter

import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"testing"
)

func TestAdapter_adapter(t *testing.T) {
	constLabels := []attribute.KeyValue{
		attribute.String("constLabels1", "constValues1"),
		attribute.Int("constLabels2", 2),
	}
	baseAdapterManager := createBaseAdapterManager(constLabels)
	attrs, _ := baseAdapterManager.detailTopologyAdapter.adapter(makeOriginGaugeGroup(300000))
	for i := 0; i < len(attrs); i++ {
		fmt.Printf("%v\n", attrs[i])
	}
}

func makeOriginGaugeGroup(latency int64) *model.GaugeGroup {
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
