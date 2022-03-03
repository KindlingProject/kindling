package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"strconv"
)

type ProtocolType string

const (
	http  = "http"
	http2 = "http2"
	grpc  = "grpc"
	dubbo = "dubbo"
	dns   = "dns"
	kafka = "kafka"
	mysql = "mysql"
)

func fillSpecialProtocolLabels(g *gauges, protocol ProtocolType) {
	switch protocol {
	case kafka:
		fillKafkaMetricProtocolLabel(g)
	default:
		// Do nothing
	}
}

func fillSpanProtocolLabels(g *gauges, protocol ProtocolType) {
	switch protocol {
	case http:
		fillSpanHttpProtocolLabel(g)
	case dns:
		fillSpanDNSProtocolLabel(g)
	case mysql:
		fillSpanMysqlProtocolLabel(g)
	}
}

func fillSpanMysqlProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue("mysql.sql", g.Labels.GetStringValue(constlabels.Sql))
	g.targetLabels.AddStringValue("mysql.error_code", g.Labels.GetStringValue(constlabels.SqlErrCode))
	g.targetLabels.AddStringValue("mysql.error_msg", g.Labels.GetStringValue(constlabels.SqlErrMsg))
}

func fillSpanDNSProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue("dns.domain", g.Labels.GetStringValue(constlabels.DnsDomain))
	g.targetLabels.AddStringValue("dns.rcode", g.Labels.GetStringValue(constlabels.DnsRcode))
}

func fillCommonProtocolLabels(g *gauges, protocol ProtocolType, isServer bool) {
	switch protocol {
	case http:
		if isServer {
			fillEntityHttpProtocolLabel(g)
		} else {
			fillTopologyHttpProtocolLabel(g)
		}
	case dns:
		if isServer {
			fillEntityDnsProtocolLabel(g)
		} else {
			fillTopologyDnsProtocolLabel(g)
		}
	case kafka:
		if isServer {
			fillEntityKafkaProtocolLabel(g)
		} else {
			fillTopologyKafkaProtocolLabel(g)
		}
	case mysql:
		if isServer {
			fillEntityMysqlProtocolLabel(g)
		} else {
			fillTopologyMysqlProtocolLabel(g)
		}
	case grpc:
		if isServer {
			fillEntityHttpProtocolLabel(g)
		} else {
			fillTopologyHttpProtocolLabel(g)
		}
	default:
		// Do nothing ?
	}
}

func fillEntityHttpProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.RequestContent, g.Labels.GetStringValue(constlabels.ContentKey))
	g.targetLabels.AddStringValue(constlabels.ResponseContent, strconv.FormatInt(g.Labels.GetIntValue(constlabels.HttpStatusCode), 10))
}

func fillTopologyHttpProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.StatusCode, strconv.FormatInt(g.Labels.GetIntValue(constlabels.HttpStatusCode), 10))
}

func fillSpanHttpProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue("http.method", g.Labels.GetStringValue(constlabels.HttpMethod))
	g.targetLabels.AddStringValue("http.endpoint", g.Labels.GetStringValue(constlabels.HttpUrl))
	g.targetLabels.AddIntValue("http.status_code", g.Labels.GetIntValue(constlabels.HttpStatusCode))
	g.targetLabels.AddStringValue("http.trace_id", g.Labels.GetStringValue(constlabels.HttpApmTraceId))
	g.targetLabels.AddStringValue("http.trace_type", g.Labels.GetStringValue(constlabels.HttpApmTraceType))
	g.targetLabels.AddStringValue("http.request_headers", g.Labels.GetStringValue(constlabels.HttpRequestPayload))
	g.targetLabels.AddStringValue("http.request_body", "")
	g.targetLabels.AddStringValue("http.response_headers", g.Labels.GetStringValue(constlabels.HttpResponsePayload))
	g.targetLabels.AddStringValue("http.response_body", "")
}

func fillEntityDnsProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.RequestContent, g.Labels.GetStringValue(constlabels.DnsDomain))
	g.targetLabels.AddStringValue(constlabels.ResponseContent, strconv.FormatInt(g.Labels.GetIntValue(constlabels.DnsRcode), 10))
}

func fillTopologyDnsProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.StatusCode, strconv.FormatInt(g.Labels.GetIntValue(constlabels.DnsRcode), 10))
}

func fillEntityKafkaProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.RequestContent, g.Labels.GetStringValue(constlabels.KafkaTopic))
	g.targetLabels.AddStringValue(constlabels.ResponseContent, g.Labels.GetStringValue(constlabels.STR_EMPTY))
}

func fillTopologyKafkaProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.StatusCode, g.Labels.GetStringValue(constlabels.KafkaTopic))
}

func fillEntityMysqlProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.RequestContent, g.Labels.GetStringValue(constlabels.ContentKey))
	g.targetLabels.AddStringValue(constlabels.ResponseContent, g.Labels.GetStringValue(constlabels.STR_EMPTY))
}

func fillTopologyMysqlProtocolLabel(g *gauges) {
	g.targetLabels.AddStringValue(constlabels.StatusCode, g.Labels.GetStringValue(constlabels.STR_EMPTY))
}

func fillKafkaMetricProtocolLabel(g *gauges) {
	// TODO Missing Information Element
	g.targetLabels.AddStringValue(constlabels.Topic, g.Labels.GetStringValue(constlabels.KafkaTopic))
	//g.targetLabels.AddStringValue(constlabels.Operation,g.Labels.GetStringValue())
	//g.targetLabels.AddStringValue(constlabels.ConsumerId, g.Labels.GetStringValue())
}
