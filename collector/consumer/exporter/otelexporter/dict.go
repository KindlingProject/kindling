package otelexporter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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
)

type valueType int

const (
	Int64 valueType = iota
	String
	Bool
	StrEmpty
	FromInt64ToString
)

type dictionary struct {
	newKey    string
	originKey string
	valueType
}

var instanceMetricDicList = []dictionary{
	{constlabels.SrcIp, constlabels.SrcIp, String},
	{constlabels.DstIp, constlabels.DstIp, String},
	{constlabels.DstPort, constlabels.DstPort, Int64},
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
	{constlabels.DstNode, constlabels.DstNode, String},
	{constlabels.DstPod, constlabels.DstPod, String},
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

func RemoveDstPodInfoForNonExternalAggTopology(labels *model.AttributeMap, attrs []attribute.KeyValue) []attribute.KeyValue {
	if constlabels.IsNamespaceNotFound(labels.GetStringValue(constlabels.DstNamespace)) {
		return attrs
	} else {
		for i := 0; i < len(attrs); i++ {
			if attrs[i].Key == constlabels.DstNode || attrs[i].Key == constlabels.DstPod {
				attrs[i].Value = attribute.StringValue(constlabels.STR_EMPTY)
			}
		}
		return attrs
	}
}

var entityProtocol = []extraLabelsParam{
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.ResponseContent, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{HTTP}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.KafkaTopic, String},
		{constlabels.RequestContent, constlabels.STR_EMPTY, StrEmpty},
	}, extraLabelsKey{KAFKA}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.RequestContent, constlabels.SqlErrCode, FromInt64ToString},
	}, extraLabelsKey{MYSQL}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.RequestContent, constlabels.HttpStatusCode, FromInt64ToString},
	}, extraLabelsKey{GRPC}},
	{[]dictionary{
		{constlabels.RequestContent, constlabels.DnsDomain, String},
		{constlabels.RequestContent, constlabels.DnsRcode, FromInt64ToString},
	}, extraLabelsKey{DNS}},
}

//
//case http:
//if isServer {
//fillEntityHttpProtocolLabel(g)
//} else {
//fillTopologyHttpProtocolLabel(g)
//}
//case dns:
//if isServer {
//fillEntityDnsProtocolLabel(g)
//} else {
//fillTopologyDnsProtocolLabel(g)
//}
//case kafka:
//if isServer {
//fillEntityKafkaProtocolLabel(g)
//} else {
//fillTopologyKafkaProtocolLabel(g)
//}
//case mysql:
//if isServer {
//fillEntityMysqlProtocolLabel(g)
//} else {
//fillTopologyMysqlProtocolLabel(g)
//}
//case grpc:
//if isServer {
//fillEntityHttpProtocolLabel(g)
//} else {
//fillTopologyHttpProtocolLabel(g)
//}
