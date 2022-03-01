package kindlingformatprocessor

import (
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"strconv"
)

type ProtocolType string

const (
	generic = "generic"
	http    = "http"
	http2   = "http2"
	grpc    = "grpc"
	dubbo   = "dubbo"
	dns     = "dns"
	kafka   = "kafka"
	mysql   = "mysql"
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
	}
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
	// TODO trace_id and trace_type
	//kv = append(kv, attribute.String("http.trace_id", s.Protocol.HTTP.TraceID))
	//kv = append(kv, attribute.String("http.trace_type", s.Protocol.HTTP.TraceType))
	g.targetLabels.AddStringValue("http.request_payload", g.Labels.GetStringValue(constlabels.HttpRequestPayload))
	g.targetLabels.AddStringValue("http.response_payload", g.Labels.GetStringValue(constlabels.HttpResponsePayload))
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
