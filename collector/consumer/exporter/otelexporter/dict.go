package otelexporter

import (
	"github.com/Kindling-project/kindling/collector/model/constlabels"
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

var entityMetricDicList = []dictionary{
	{constlabels.DstNamespace, constlabels.Namespace, String},
	{constlabels.WorkloadKind, constlabels.DstWorkloadKind, String},
	{constlabels.WorkloadName, constlabels.DstWorkloadName, String},
	{constlabels.Service, constlabels.DstService, String},
	{constlabels.Protocol, constlabels.Protocol, String},
}

var entityProtocol = []extraLabelsParam{
	{[]dictionary{
		{constlabels.RequestContent, constlabels.ContentKey, String},
		{constlabels.ResponseContent, constlabels.HttpStatusCode, Int64},
	}, extraLabelsKey{HTTP}},
	//{[]dictionary{
	//	{constlabels.RequestContent, constlabels.ContentKey, String},
	//	{constlabels.RequestContent, constlabels.HttpStatusCode, Int64},
	//}, extraLabelsKey{KAFKA}},
	//{[]dictionary{
	//	{constlabels.RequestContent, constlabels.ContentKey, String},
	//	{constlabels.RequestContent, constlabels.HttpStatusCode, Int64},
	//}, extraLabelsKey{MYSQL}},
}
