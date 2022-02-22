package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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
	targetValues []model.Gauge
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
		targetValues: make([]model.Gauge, 0, len(g.Values)),
		targetLabels: model.NewAttributeMap(),
	}
}

func (g *gauges) isSlowOrError() bool {
	return g.Labels.GetBoolValue(constlabels.IsSlow) || g.Labels.GetBoolValue(constlabels.IsError)
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
		if name := constlabels.ToKindlingMetricName(gauge.Name, g.Labels.GetBoolValue(constlabels.IsServer)); name != "" {
			g.targetValues = append(g.targetValues, model.Gauge{
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

	g.targetValues = append(g.targetValues, model.Gauge{
		Name:  constlabels.ToKindlingTraceAsMetricName(),
		Value: requestDuration,
	})
}

func ProtocolDetailMetricName(cfg *Config, g *gauges) {
	for _, gauge := range g.Values {
		g.targetValues = append(g.targetValues, model.Gauge{
			Name:  constlabels.ToKindlingDetailMetricName(gauge.Name, g.Labels.GetStringValue(constlabels.Protocol)),
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

func SrcDockerInfo(cfg *Config, g *gauges) {
	if cfg.NeedPodDetail {
		g.targetLabels.AddStringValue(constlabels.SrcContainer, g.Labels.GetStringValue(constlabels.SrcContainer))
		g.targetLabels.AddStringValue(constlabels.SrcContainerId, g.Labels.GetStringValue(constlabels.SrcContainerId))
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
