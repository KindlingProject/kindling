package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
)

type Protocol int

const (
	empty Protocol = iota
	HTTP
	KAFKA
	DNS
	MYSQL
	GRPC
	UNSUPPORTED
)

type valueType int

const (
	Int64 valueType = iota
	String
	Bool
	StrEmpty
	FromInt64ToString
)

const (
	millToNano   = 1000000
	GreenStatus  = "1"
	YellowStatus = "2"
	RedStatus    = "3"
)

type dictionary struct {
	newKey    string
	originKey string
	valueType
}

var isSlowDicList = []dictionary{
	{constlabels.IsSlow, constlabels.IsSlow, Bool},
}

var topologyInstanceMetricDicList = []dictionary{
	{constlabels.SrcIp, constlabels.SrcIp, String},
	{constlabels.DstIp, constlabels.DstIp, String},
	{constlabels.DstPort, constlabels.DstPort, FromInt64ToString},
}

var entityInstanceMetricDicList = []dictionary{
	{constlabels.Ip, constlabels.DstIp, String},
	{constlabels.Port, constlabels.DstPort, FromInt64ToString},
}

var entityDetailMetricDicList = []dictionary{
	{constlabels.Node, constlabels.DstNode, String},
	{constlabels.Pod, constlabels.DstPod, String},
	{constlabels.Container, constlabels.DstContainer, String},
	{constlabels.ContainerId, constlabels.DstContainerId, String},
}

var entityMetricDicList = []dictionary{
	{constlabels.Namespace, constlabels.DstNamespace, String},
	{constlabels.WorkloadKind, constlabels.DstWorkloadKind, String},
	{constlabels.WorkloadName, constlabels.DstWorkloadName, String},
	{constlabels.Service, constlabels.DstService, String},
	{constlabels.Protocol, constlabels.Protocol, String},
}

var topologyDetailMetricDicList = []dictionary{
	{constlabels.SrcContainerId, constlabels.SrcContainerId, String},
	{constlabels.SrcContainer, constlabels.SrcContainer, String},
	{constlabels.DstContainerId, constlabels.DstContainerId, String},
	{constlabels.DstContainer, constlabels.DstContainer, String},
	// this info has contained in topology baseInfos
	//{constlabels.DstNode, constlabels.DstNode, String},
	//{constlabels.DstPod, constlabels.DstPod, String},
	{constlabels.SrcNode, constlabels.SrcNode, String},
	{constlabels.SrcPod, constlabels.SrcPod, String},
}

var SpanDicList = []dictionary{
	{constlabels.SpanSrcContainerId, constlabels.SrcContainerId, String},
	{constlabels.SpanSrcContainerName, constlabels.SrcContainer, String},
	{constlabels.SpanDstContainerId, constlabels.DstContainerId, String},
	{constlabels.SpanDstContainerName, constlabels.DstContainer, String},
	// this info has contained in topology baseInfos
	//{constlabels.DstNode, constlabels.DstNode, String},
	//{constlabels.DstPod, constlabels.DstPod, String},
	{constlabels.SrcNode, constlabels.SrcNode, String},
	{constlabels.SrcPod, constlabels.SrcPod, String},
}

var topologyMetricDicList = []dictionary{
	{constlabels.SrcNamespace, constlabels.SrcNamespace, String},
	{constlabels.SrcWorkloadKind, constlabels.SrcWorkloadKind, String},
	{constlabels.SrcWorkloadName, constlabels.SrcWorkloadName, String},
	{constlabels.SrcService, constlabels.SrcService, String},

	{constlabels.DstNamespace, constlabels.DstNamespace, String},
	{constlabels.DstWorkloadKind, constlabels.DstWorkloadKind, String},
	{constlabels.DstWorkloadName, constlabels.DstWorkloadName, String},
	{constlabels.DstService, constlabels.DstService, String},

	{constlabels.DstNode, constlabels.DstNode, String},
	{constlabels.DstPod, constlabels.DstPod, String},

	{constlabels.Protocol, constlabels.Protocol, String},
}

func removeDstPodInfoForNonExternal() adjustFunctions {
	return adjustFunctions{
		adjustAttrMaps: func(labels *model.AttributeMap, attributeMap *model.AttributeMap) *model.AttributeMap {
			if constlabels.IsNamespaceNotFound(labels.GetStringValue(constlabels.DstNamespace)) {
				return attributeMap
			} else {
				attributeMap.AddStringValue(constlabels.DstNode, constlabels.STR_EMPTY)
				attributeMap.AddStringValue(constlabels.DstPod, constlabels.STR_EMPTY)
				return attributeMap
			}
		},
		adjustLabels: func(labels *model.AttributeMap, attrs []attribute.KeyValue) []attribute.KeyValue {
			if constlabels.IsNamespaceNotFound(labels.GetStringValue(constlabels.DstNamespace)) {
				return attrs
			} else {
				for i := 0; i < len(attrs); i++ {
					if attrs[i].Key == constlabels.DstNode || attrs[i].Key == constlabels.DstPod {
						attrs[i].Value = attribute.StringValue(constlabels.STR_EMPTY)
					}
				}
			}
			return attrs
		},
	}
}

func replaceDstIpOrDstPortByDNat() adjustFunctions {
	return adjustFunctions{
		adjustAttrMaps: func(labels *model.AttributeMap, attributeMap *model.AttributeMap) *model.AttributeMap {
			dNatIp := labels.GetStringValue(constlabels.DnatIp)
			dNatPort := labels.GetIntValue(constlabels.DnatPort)
			if dNatIp == "" || dNatPort < 1 {
				return attributeMap
			} else {
				attributeMap.AddStringValue(constlabels.DstIp, dNatIp)
				attributeMap.AddIntValue(constlabels.DstPort, dNatPort)
				return attributeMap
			}
		},
		adjustLabels: func(labels *model.AttributeMap, attrs []attribute.KeyValue) []attribute.KeyValue {
			dNatIp := labels.GetStringValue(constlabels.DnatIp)
			dNatPort := labels.GetIntValue(constlabels.DnatPort)
			if dNatIp == "" || dNatPort < 1 {
				return attrs
			} else {
				for i := 0; i < len(attrs); i++ {
					if attrs[i].Key == constlabels.DstIp && dNatIp != "" {
						attrs[i].Value = attribute.StringValue(dNatIp)
					} else if attrs[i].Key == constlabels.DstPort && dNatPort > 0 {
						attrs[i].Value = attribute.Int64Value(dNatPort)
					}
				}
			}
			return attrs
		},
	}
}

var entityProtocol = []extraLabelsParam{
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.ResponseContent, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{HTTP}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.KafkaTopic, String},
		{constlabels.ResponseContent, constlabels.STR_EMPTY, StrEmpty},
	}, extraLabelsKey{KAFKA}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.ResponseContent, constlabels.SqlErrCode, FromInt64ToString},
	}, extraLabelsKey{MYSQL}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.ResponseContent, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{GRPC}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.DnsDomain, String},
		{constlabels.ResponseContent, constlabels.DnsRcode, FromInt64ToString},
	}, extraLabelsKey{DNS}},
	{
		[]dictionary{}, extraLabelsKey{UNSUPPORTED},
	},
}

var spanProtocol = []extraLabelsParam{
	{[]dictionary{
		{constlabels.SpanHttpMethod, constlabels.HttpMethod, String},
		{constlabels.SpanHttpEndpoint, constlabels.HttpUrl, String},
		{constlabels.SpanHttpStatusCode, constlabels.HttpStatusCode, Int64},
		{constlabels.SpanHttpTraceId, constlabels.HttpApmTraceId, String},
		{constlabels.SpanHttpTraceType, constlabels.HttpApmTraceType, String},
		{constlabels.SpanHttpRequestHeaders, constlabels.HttpRequestPayload, String},
		{constlabels.SpanHttpRequestBody, constlabels.STR_EMPTY, StrEmpty},
		{constlabels.SpanHttpResponseHeaders, constlabels.HttpResponsePayload, String},
		{constlabels.SpanHttpResponseBody, constlabels.STR_EMPTY, StrEmpty},
	}, extraLabelsKey{HTTP}},
	{[]dictionary{
		{constlabels.SpanMysqlSql, constlabels.Sql, String},
		{constlabels.SpanMysqlErrorCode, constlabels.SqlErrCode, Int64},
		{constlabels.SpanMysqlErrorMsg, constlabels.SqlErrMsg, String},
	}, extraLabelsKey{MYSQL}},
	{[]dictionary{
		{constlabels.SpanDnsDomain, constlabels.DnsDomain, String},
		{constlabels.SpanDnsRCode, constlabels.DnsRcode, FromInt64ToString},
	}, extraLabelsKey{DNS}},
	{
		[]dictionary{}, extraLabelsKey{UNSUPPORTED},
	},
}

var topologyProtocol = []extraLabelsParam{
	{[]dictionary{
		{constlabels.StatusCode, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{HTTP}},
	{[]dictionary{
		{constlabels.StatusCode, constlabels.STR_EMPTY, StrEmpty},
		//{constlabels.HttpStatusCode, constlabels.STR_EMPTY, StrEmpty},
	}, extraLabelsKey{KAFKA}},
	{[]dictionary{
		{constlabels.StatusCode, constlabels.SqlErrCode, FromInt64ToString},
	}, extraLabelsKey{MYSQL}},
	{[]dictionary{
		{constlabels.StatusCode, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{GRPC}},
	{[]dictionary{
		{constlabels.StatusCode, constlabels.DnsRcode, FromInt64ToString},
	}, extraLabelsKey{DNS}},
	{
		[]dictionary{}, extraLabelsKey{UNSUPPORTED},
	},
}

var traceSpanStatus = []dictionary{
	{constlabels.RequestSentNs, constlabels.STR_EMPTY, Int64},
	{constlabels.WaitingTTfbNs, constlabels.STR_EMPTY, Int64},
	{constlabels.ContentDownloadNs, constlabels.STR_EMPTY, Int64},
	{constlabels.RequestTotalNs, constlabels.STR_EMPTY, Int64},
	{constlabels.RequestIoBytes, constlabels.STR_EMPTY, Int64},
	{constlabels.ResponseIoBytes, constlabels.STR_EMPTY, Int64},
	{constlabels.IsServer, constlabels.STR_EMPTY, Int64},
	{constlabels.IsError, constlabels.STR_EMPTY, Int64},
	{constlabels.IsSlow, constlabels.STR_EMPTY, Int64},
	{constlabels.IsConvergent, constlabels.STR_EMPTY, Int64},
	{constlabels.Timestamp, constlabels.STR_EMPTY, Int64},
}

func getTraceSpanStatusLabels(gaugeGroup *model.GaugeGroup) []attribute.KeyValue {
	valueLabels := make([]attribute.KeyValue, 11)
	for i := 0; i < len(gaugeGroup.Values); i++ {
		switch gaugeGroup.Values[i].Name {
		case constvalues.RequestSentTime:
			valueLabels[0] = attribute.Int64(traceSpanStatus[0].newKey, gaugeGroup.Values[i].Value)
		case constvalues.WaitingTtfbTime:
			valueLabels[1] = attribute.Int64(traceSpanStatus[1].newKey, gaugeGroup.Values[i].Value)
		case constvalues.ContentDownloadTime:
			valueLabels[2] = attribute.Int64(traceSpanStatus[2].newKey, gaugeGroup.Values[i].Value)
		case constvalues.RequestTotalTime:
			valueLabels[3] = attribute.Int64(traceSpanStatus[3].newKey, gaugeGroup.Values[i].Value)
		case constvalues.RequestIo:
			valueLabels[4] = attribute.Int64(traceSpanStatus[4].newKey, gaugeGroup.Values[i].Value)
		case constvalues.ResponseIo:
			valueLabels[5] = attribute.Int64(traceSpanStatus[5].newKey, gaugeGroup.Values[i].Value)
		}
	}

	valueLabels[6] = attribute.Int64(traceSpanStatus[6].newKey, int64(If(gaugeGroup.Labels.GetBoolValue(constlabels.IsServer), 1, 0).(int)))
	valueLabels[7] = attribute.Int64(traceSpanStatus[7].newKey, int64(If(gaugeGroup.Labels.GetBoolValue(constlabels.IsError), 1, 0).(int)))
	valueLabels[8] = attribute.Int64(traceSpanStatus[8].newKey, int64(If(gaugeGroup.Labels.GetBoolValue(constlabels.IsSlow), 1, 0).(int)))
	valueLabels[9] = attribute.Int64(traceSpanStatus[9].newKey, 0)
	valueLabels[10] = attribute.Int64(traceSpanStatus[10].newKey, int64(gaugeGroup.Timestamp/millToNano))
	return valueLabels
}

var traceStatus = []dictionary{
	{constlabels.RequestReqxferStatus, constlabels.STR_EMPTY, String},
	{constlabels.RequestProcessingStatus, constlabels.STR_EMPTY, String},
	{constlabels.ResponseRspxferStatus, constlabels.STR_EMPTY, String},
	{constlabels.RequestDurationStatus, constlabels.STR_EMPTY, String},
	{constlabels.IsServer, constlabels.STR_EMPTY, Bool},
}

func getTraceStatusLabels(gaugeGroup *model.GaugeGroup) []attribute.KeyValue {
	var requestSend, waitingTtfb, contentDownload, requestTotalTime int64
	for i := 0; i < len(gaugeGroup.Values); i++ {
		if gaugeGroup.Values[i].Name == constvalues.RequestSentTime {
			requestSend = gaugeGroup.Values[i].Value
		} else if gaugeGroup.Values[i].Name == constvalues.WaitingTtfbTime {
			waitingTtfb = gaugeGroup.Values[i].Value
		} else if gaugeGroup.Values[i].Name == constvalues.ContentDownloadTime {
			contentDownload = gaugeGroup.Values[i].Value
		} else if gaugeGroup.Values[i].Name == constvalues.RequestTotalTime {
			requestTotalTime = gaugeGroup.Values[i].Value
		}
	}

	return []attribute.KeyValue{
		attribute.String(traceStatus[0].newKey, getSubStageStatus(requestSend)),
		attribute.String(traceStatus[1].newKey, getSubStageStatus(waitingTtfb)),
		attribute.String(traceStatus[2].newKey, getSubStageStatus(contentDownload)),
		attribute.String(traceStatus[3].newKey, getRequestStatus(requestTotalTime)),
		attribute.Bool(traceStatus[4].newKey, gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)),
	}
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

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
