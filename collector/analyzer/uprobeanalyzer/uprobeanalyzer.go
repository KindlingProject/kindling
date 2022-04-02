package uprobeanalyzer

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/analyzer/tools"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	conntrackerpackge "github.com/Kindling-project/kindling/collector/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

var httpMerger = tools.NewHttpMergeCache()

const (
	UprobeType analyzer.Type = "uprobeanalyzer"

	clientRole = 1
	serverRole = 2

	MillisecondToNanosecond = 1e6
)

type UprobeAnalyzer struct {
	cfg         *Config
	consumers   []consumer.Consumer
	conntracker *conntrackerpackge.Conntracker

	telemetry *component.TelemetryTools
}

func NewUprobeAnalyzer(config interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Sugar().Panicf("Cannot convert mock_analyzer config")
	}
	retAnalyzer := &UprobeAnalyzer{
		cfg:       cfg,
		consumers: consumers,
		telemetry: telemetry,
	}
	conntracker, err := conntrackerpackge.NewConntracker(10000)
	if err != nil {
		telemetry.Logger.Panic("Failed to create UprobeAnalyzer: ", zap.Error(err))
	}
	retAnalyzer.conntracker = conntracker
	return retAnalyzer
}

func (a *UprobeAnalyzer) Start() error {
	return nil
}

func (a *UprobeAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	if event.Name != constnames.GrpcUprobeEvent {
		a.telemetry.Logger.Warn("Skip a Event: UprobeAnalyzer cannot handle event", zap.String("eventName", event.Name))
		return nil
	}

	role := event.GetIntUserAttribute("trace_role")
	pid := event.GetIntUserAttribute("pid")
	fd := event.GetIntUserAttribute("fd")
	if role != clientRole && role != serverRole {
		a.telemetry.Logger.Warn("Skip a Event: UprobeAnalyzer received a unexpected role event", zap.Int64("pid", pid), zap.Int64("fd", fd))
		return nil
	}

	remoteIp := event.GetStringUserAttribute("remote_addr")
	remotePort := event.GetIntUserAttribute("remote_port")
	containerId := event.GetStringUserAttribute("containerid")
	reqMethod := event.GetStringUserAttribute("req_method")
	reqPath := event.GetStringUserAttribute("req_path")
	statusCode := event.GetIntUserAttribute("resp_status")
	reqBody := event.GetUserAttribute("req_body").GetValue()
	respBody := event.GetUserAttribute("resp_body").GetValue()

	// unit: nanosecond
	latency := event.GetLatency()
	isSlow := latency >= a.cfg.ResponseSlowThreshold*MillisecondToNanosecond

	labels := model.NewAttributeMapWithValues(map[string]model.AttributeValue{
		constlabels.Pid:                 model.NewIntValue(pid),
		constlabels.DnatIp:              model.NewStringValue(constlabels.STR_EMPTY),
		constlabels.DnatPort:            model.NewIntValue(-1),
		constlabels.ContainerId:         model.NewStringValue(containerId),
		constlabels.IsSlow:              model.NewBoolValue(isSlow),
		constlabels.Protocol:            model.NewStringValue("grpc"),
		constlabels.HttpMethod:          model.NewStringValue(reqMethod),
		constlabels.HttpUrl:             model.NewStringValue(reqPath),
		constlabels.HttpRequestPayload:  model.NewStringValue(string(reqBody)),
		constlabels.HttpResponsePayload: model.NewStringValue(string(respBody)),
		constlabels.HttpStatusCode:      model.NewIntValue(statusCode),
		constlabels.ContentKey:          model.NewStringValue(getContentKey(reqPath)),
	})

	if statusCode >= 400 {
		labels.AddBoolValue(constlabels.IsError, true)
		labels.AddIntValue(constlabels.ErrorType, int64(constlabels.ProtocolError))
	}

	if role == clientRole {
		srcIp := event.GetUserAttribute("src_ip")
		srcPort := event.GetUserAttribute("src_port")
		var srcIpValue string
		var srcPortValue int64
		if srcIp != nil && srcPort != nil {
			srcIpValue = string(srcIp.Value)
			srcPortValue = event.GetIntUserAttribute("src_port")
		}
		labels.Merge(model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcIp:    model.NewStringValue(srcIpValue),
			constlabels.DstIp:    model.NewStringValue(remoteIp),
			constlabels.SrcPort:  model.NewIntValue(srcPortValue),
			constlabels.DstPort:  model.NewIntValue(remotePort),
			constlabels.IsServer: model.NewBoolValue(false),
		}))
		// Find dst NAT information
		dNatInfo := a.conntracker.GetDNATTupleWithString(srcIpValue, remoteIp, uint16(srcPortValue), uint16(remotePort), 0)
		if dNatInfo != nil {
			labels.AddStringValue(constlabels.DnatIp, dNatInfo.ReplSrcIP.String())
			labels.AddIntValue(constlabels.DnatPort, int64(dNatInfo.ReplSrcPort))
		}
	} else if role == serverRole {
		labels.Merge(model.NewAttributeMapWithValues(map[string]model.AttributeValue{
			constlabels.SrcIp:    model.NewStringValue(remoteIp),
			constlabels.DstIp:    model.NewStringValue(""),
			constlabels.SrcPort:  model.NewIntValue(remotePort),
			constlabels.DstPort:  model.NewIntValue(-1),
			constlabels.IsServer: model.NewBoolValue(true),
		}))
	}

	latencyGauge := &model.Gauge{
		Name:  constvalues.RequestTotalTime,
		Value: int64(latency),
	}
	requestIoGauge := &model.Gauge{
		Name:  constvalues.RequestIo,
		Value: event.GetIntUserAttribute("req_body_size"),
	}
	responseIoGauge := &model.Gauge{
		Name:  constvalues.ResponseIo,
		Value: event.GetIntUserAttribute("resp_body_size"),
	}
	gaugeGroup := model.NewGaugeGroup("GrpcUprobeGroup", labels, event.Timestamp,
		latencyGauge, requestIoGauge, responseIoGauge)
	var retError error
	for _, nextConsumer := range a.consumers {
		err := nextConsumer.Consume(gaugeGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	return retError
}

func (a *UprobeAnalyzer) Shutdown() error {
	return nil
}

func (a *UprobeAnalyzer) Type() analyzer.Type {
	return UprobeType
}

func getContentKey(url string) string {
	if url == "" {
		return ""
	}
	index := strings.Index(url, "?")
	if index != -1 {
		url = url[:index]
	}
	return httpMerger.GetContentKey(url)
}
