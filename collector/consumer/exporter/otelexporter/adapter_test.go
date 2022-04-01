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
	detailTopologyAttrs, _ := baseAdapterManager.detailTopologyAdapter.adapter(makeOriginGaugeGroup(300000))
	for i := 0; i < len(detailTopologyAttrs); i++ {
		fmt.Printf("%+v\n", detailTopologyAttrs[i])
	}

	detailTraceSpanAttrs, _ := baseAdapterManager.traceToSpanAdapter.adapter(makeOriginGaugeGroup(300000))
	for i := 0; i < len(detailTopologyAttrs); i++ {
		fmt.Printf("%+v\n", detailTraceSpanAttrs[i])
	}

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
