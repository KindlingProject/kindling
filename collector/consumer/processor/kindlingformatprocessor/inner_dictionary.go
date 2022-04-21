package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
)

const (
	millToNano   = 1000000
	GreenStatus  = "1"
	YellowStatus = "2"
	RedStatus    = "3"
)

type gauges struct {
	*model.GaugeGroup
	targetValues []*model.Gauge
	targetLabels *model.AttributeMap
}

func newGauges(g *model.GaugeGroup) *gauges {
	// Since some operations modify the original data, a copy is used here instead of the original data (only shallow copy)
	gaugeGroupCp := &model.GaugeGroup{
		Name:      g.Name,
		Values:    g.Values,
		Labels:    g.Labels,
		Timestamp: g.Timestamp,
	}

	return &gauges{
		GaugeGroup:   gaugeGroupCp,
		targetValues: make([]*model.Gauge, 0, len(g.Values)),
		targetLabels: model.NewAttributeMap(),
	}
}

func (g gauges) getResult() *model.GaugeGroup {
	return &model.GaugeGroup{
		Name:      g.Name,
		Values:    g.targetValues,
		Labels:    g.targetLabels,
		Timestamp: g.Timestamp,
	}
}

type Relabel func(cfg *Config, g *gauges)

func (g gauges) Process(cfg *Config, relabels ...Relabel) *model.GaugeGroup {
	for i := 0; i < len(relabels); i++ {
		relabels[i](cfg, &g)
	}
	return g.getResult()
}

func MetricName(cfg *Config, g *gauges) {
	for _, gauge := range g.Values {
		if name := constnames.ToKindlingNetMetricName(gauge.Name, g.Labels.GetBoolValue(constlabels.IsServer)); name != "" {
			g.targetValues = append(g.targetValues, &model.Gauge{
				Name:  name,
				Value: gauge.Value,
			})
			//} else {
			//	log.Debugf(nil, "Unused Metric Value：%s,dropped!", gauge.Name)
		}
	}
}

func TraceName(cfg *Config, g *gauges) {
	var requestDuration int64
	for i := 0; i < len(g.Values); i++ {
		if g.Values[i].Name == constvalues.RequestTotalTime {
			requestDuration = g.Values[i].Value
		}
	}

	g.targetValues = append(g.targetValues, &model.Gauge{
		Name:  constnames.TraceAsMetric,
		Value: requestDuration,
	})
}

func SpanName(cfg *Config, g *gauges) {
	g.Name = constvalues.SpanInfo
}

func ProtocolDetailMetricName(cfg *Config, g *gauges) {
	for _, gauge := range g.Values {
		g.targetValues = append(g.targetValues, &model.Gauge{
			Name:  constnames.ToKindlingDetailMetricName(gauge.Name, g.Labels.GetStringValue(constlabels.Protocol)),
			Value: gauge.Value,
		})
	}
}

func ServiceK8sInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		g.targetLabels.AddStringValue(constlabels.Node, g.Labels.GetStringValue(constlabels.DstNode))
		g.targetLabels.AddStringValue(constlabels.Pod, g.Labels.GetStringValue(constlabels.DstPod))
		g.targetLabels.AddStringValue(constlabels.Container, g.Labels.GetStringValue(constlabels.DstContainer))
		g.targetLabels.AddStringValue(constlabels.ContainerId, g.Labels.GetStringValue(constlabels.DstContainerId))
	}
	g.targetLabels.AddStringValue(constlabels.Namespace, g.Labels.GetStringValue(constlabels.DstNamespace))
	g.targetLabels.AddStringValue(constlabels.WorkloadKind, g.Labels.GetStringValue(constlabels.DstWorkloadKind))
	g.targetLabels.AddStringValue(constlabels.WorkloadName, g.Labels.GetStringValue(constlabels.DstWorkloadName))
	g.targetLabels.AddStringValue(constlabels.Service, g.Labels.GetStringValue(constlabels.DstService))
}

func TopologyTraceK8sInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.SrcNode, g.Labels.GetStringValue(constlabels.SrcNode))
	g.targetLabels.AddStringValue(constlabels.SrcNamespace, g.Labels.GetStringValue(constlabels.SrcNamespace))
	g.targetLabels.AddStringValue(constlabels.SrcWorkloadKind, g.Labels.GetStringValue(constlabels.SrcWorkloadKind))
	g.targetLabels.AddStringValue(constlabels.SrcWorkloadName, g.Labels.GetStringValue(constlabels.SrcWorkloadName))
	g.targetLabels.AddStringValue(constlabels.SrcService, g.Labels.GetStringValue(constlabels.SrcService))
	g.targetLabels.AddStringValue(constlabels.SrcPod, g.Labels.GetStringValue(constlabels.SrcPod))

	g.targetLabels.AddStringValue(constlabels.DstNode, g.Labels.GetStringValue(constlabels.DstNode))
	g.targetLabels.AddStringValue(constlabels.DstNamespace, g.Labels.GetStringValue(constlabels.DstNamespace))
	g.targetLabels.AddStringValue(constlabels.DstWorkloadKind, g.Labels.GetStringValue(constlabels.DstWorkloadKind))
	g.targetLabels.AddStringValue(constlabels.DstWorkloadName, g.Labels.GetStringValue(constlabels.DstWorkloadName))
	g.targetLabels.AddStringValue(constlabels.DstService, g.Labels.GetStringValue(constlabels.DstService))
	g.targetLabels.AddStringValue(constlabels.DstPod, g.Labels.GetStringValue(constlabels.DstPod))
}

func TopologyK8sInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		TopologyTraceK8sInfo(cfg, g)
	} else {
		g.targetLabels.AddStringValue(constlabels.SrcNamespace, g.Labels.GetStringValue(constlabels.SrcNamespace))
		g.targetLabels.AddStringValue(constlabels.SrcWorkloadKind, g.Labels.GetStringValue(constlabels.SrcWorkloadKind))
		g.targetLabels.AddStringValue(constlabels.SrcWorkloadName, g.Labels.GetStringValue(constlabels.SrcWorkloadName))
		g.targetLabels.AddStringValue(constlabels.SrcService, g.Labels.GetStringValue(constlabels.SrcService))

		g.targetLabels.AddStringValue(constlabels.DstNamespace, g.Labels.GetStringValue(constlabels.DstNamespace))
		g.targetLabels.AddStringValue(constlabels.DstWorkloadKind, g.Labels.GetStringValue(constlabels.DstWorkloadKind))
		g.targetLabels.AddStringValue(constlabels.DstWorkloadName, g.Labels.GetStringValue(constlabels.DstWorkloadName))
		g.targetLabels.AddStringValue(constlabels.DstService, g.Labels.GetStringValue(constlabels.DstService))

		if constlabels.IsNamespaceNotFound(g.Labels.GetStringValue(constlabels.DstNamespace)) {
			g.targetLabels.AddStringValue(constlabels.DstNode, g.Labels.GetStringValue(constlabels.DstNode))
			g.targetLabels.AddStringValue(constlabels.DstPod, g.Labels.GetStringValue(constlabels.DstPod))
		}
	}
}

// SrcContainerInfo adds container level information to the input gauges if cfg.NeedPodDetail is enabled.
func SrcContainerInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		g.targetLabels.AddStringValue(constlabels.SrcContainer, g.Labels.GetStringValue(constlabels.SrcContainer))
		g.targetLabels.AddStringValue(constlabels.SrcContainerId, g.Labels.GetStringValue(constlabels.SrcContainerId))
	}
}

// DstContainerInfo adds container level information to the input gauges if cfg.NeedPodDetail is enabled.
func DstContainerInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		g.targetLabels.AddStringValue(constlabels.DstContainer, g.Labels.GetStringValue(constlabels.DstContainer))
		g.targetLabels.AddStringValue(constlabels.DstContainerId, g.Labels.GetStringValue(constlabels.DstContainerId))
	}
}

func ServiceInstanceInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		g.targetLabels.AddStringValue(constlabels.Ip, g.Labels.GetStringValue(constlabels.DstIp))
		g.targetLabels.AddIntValue(constlabels.Port, g.Labels.GetIntValue(constlabels.DstPort))
	}
}

func TopologyTraceInstanceInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.SrcIp, g.Labels.GetStringValue(constlabels.SrcIp))
	DstInstanceInfo(cfg, g)
}

func TopologyInstanceInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		TopologyTraceInstanceInfo(cfg, g)
	} else if constlabels.IsNamespaceNotFound(g.Labels.GetStringValue(constlabels.DstNamespace)) {
		DstInstanceInfo(cfg, g)
	}
}

func DstInstanceInfo(cfg *Config, g *gauges) {
	if g.Labels.GetStringValue(constlabels.DnatIp) != "" {
		g.targetLabels.AddStringValue(constlabels.DstIp, g.Labels.GetStringValue(constlabels.DnatIp))
	} else {
		g.targetLabels.AddStringValue(constlabels.DstIp, g.Labels.GetStringValue(constlabels.DstIp))
	}

	if g.Labels.GetIntValue(constlabels.DnatPort) != -1 && g.Labels.GetIntValue(constlabels.DnatPort) != 0 {
		g.targetLabels.AddIntValue(constlabels.DstPort, g.Labels.GetIntValue(constlabels.DnatPort))
	} else {
		g.targetLabels.AddIntValue(constlabels.DstPort, g.Labels.GetIntValue(constlabels.DstPort))
	}
}

func SpanProtocolInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.Protocol, g.Labels.GetStringValue(constlabels.Protocol))
	g.targetLabels.AddStringValue(constlabels.ContentKey, g.Labels.GetStringValue(constlabels.ContentKey))
	fillSpanProtocolLabels(g, ProtocolType(g.Labels.GetStringValue(constlabels.Protocol)))
}

func ServiceProtocolInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.Protocol, g.Labels.GetStringValue(constlabels.Protocol))
	fillCommonProtocolLabels(g, ProtocolType(g.Labels.GetStringValue(constlabels.Protocol)), true)
}

func TopologyProtocolInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.Protocol, g.Labels.GetStringValue(constlabels.Protocol))
	fillCommonProtocolLabels(g, ProtocolType(g.Labels.GetStringValue(constlabels.Protocol)), false)
}

func ProtocolDetailInfo(cfg *Config, g *gauges) {
	fillKafkaMetricProtocolLabel(g)
}

func AddIsSlowLabel(cfg *Config, g *gauges) {
	g.targetLabels.AddBoolValue(constlabels.IsSlow, g.Labels.GetBoolValue(constlabels.IsSlow))
}

func TraceStatusInfo(cfg *Config, g *gauges) {
	var requestSend, waitingTtfb, contentDownload, requestTotalTime int64
	for i := 0; i < len(g.Values); i++ {
		if g.Values[i].Name == constvalues.RequestSentTime {
			requestSend = g.Values[i].Value
		} else if g.Values[i].Name == constvalues.WaitingTtfbTime {
			waitingTtfb = g.Values[i].Value
		} else if g.Values[i].Name == constvalues.ContentDownloadTime {
			contentDownload = g.Values[i].Value
		} else if g.Values[i].Name == constvalues.RequestTotalTime {
			requestTotalTime = g.Values[i].Value
		}
	}
	g.targetLabels.AddStringValue(constlabels.RequestReqxferStatus, getSubStageStatus(requestSend))
	g.targetLabels.AddStringValue(constlabels.RequestProcessingStatus, getSubStageStatus(waitingTtfb))
	g.targetLabels.AddStringValue(constlabels.ResponseRspxferStatus, getSubStageStatus(contentDownload))
	g.targetLabels.AddStringValue(constlabels.RequestDurationStatus, getRequestStatus(requestTotalTime))
	g.targetLabels.AddBoolValue(constlabels.IsServer, g.Labels.GetBoolValue(constlabels.IsServer))
}

func getRequestStatus(requestLatency int64) string {
	if requestLatency <= 800*millToNano {
		return GreenStatus
	} else if requestLatency >= 1500*millToNano {
		return RedStatus
	} else {
		return YellowStatus
	}
}

func getSubStageStatus(requestSendTime int64) string {
	if requestSendTime <= 200*millToNano {
		return GreenStatus
	} else if requestSendTime >= 1000*millToNano {
		return RedStatus
	} else {
		return YellowStatus
	}
}

func traceSpanContainerInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(srcContainerName, g.Labels.GetStringValue(constlabels.SrcContainer))
	g.targetLabels.AddStringValue(srcContainerId, g.Labels.GetStringValue(constlabels.SrcContainerId))
	g.targetLabels.AddStringValue(dstContainerName, g.Labels.GetStringValue(constlabels.DstContainer))
	g.targetLabels.AddStringValue(dstContainerId, g.Labels.GetStringValue(constlabels.DstContainerId))
}

func traceSpanInstanceInfo(cfg *Config, g *gauges) {
	g.targetLabels.AddStringValue(constlabels.SrcIp, g.Labels.GetStringValue(constlabels.SrcIp))
	g.targetLabels.AddStringValue(constlabels.SrcPort, g.Labels.GetStringValue(constlabels.SrcPort))
	DstInstanceInfo(cfg, g)
}

func traceSpanValuesToLabel(cfg *Config, g *gauges) {
	for i := 0; i < len(g.Values); i++ {
		switch g.Values[i].Name {
		case constvalues.RequestSentTime:
			g.targetLabels.AddIntValue(constlabels.RequestSentNs, g.Values[i].Value)
		case constvalues.WaitingTtfbTime:
			g.targetLabels.AddIntValue(constlabels.WaitingTTfbNs, g.Values[i].Value)
		case constvalues.ContentDownloadTime:
			g.targetLabels.AddIntValue(constlabels.ContentDownloadNs, g.Values[i].Value)
		case constvalues.RequestTotalTime:
			g.targetLabels.AddIntValue(constlabels.RequestTotalNs, g.Values[i].Value)
		case constvalues.RequestIo:
			g.targetLabels.AddIntValue(constlabels.RequestIoBytes, g.Values[i].Value)
		case constvalues.ResponseIo:
			g.targetLabels.AddIntValue(constlabels.ResponseIoBytes, g.Values[i].Value)
		}
	}

	g.targetLabels.AddIntValue(constlabels.IsServer, int64(If(g.Labels.GetBoolValue(constlabels.IsServer), 1, 0).(int)))
	g.targetLabels.AddIntValue(constlabels.IsError, int64(If(g.Labels.GetBoolValue(constlabels.IsError), 1, 0).(int)))
	g.targetLabels.AddIntValue(constlabels.IsSlow, int64(If(g.Labels.GetBoolValue(constlabels.IsSlow), 1, 0).(int)))

	// TODO is_convergent
	g.targetLabels.AddIntValue(constlabels.IsConvergent, 0)
	g.targetLabels.AddIntValue(constlabels.Timestamp, int64(g.Timestamp/millToNano))
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
