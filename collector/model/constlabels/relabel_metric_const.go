package constlabels

import "github.com/Kindling-project/kindling/collector/model/constvalues"

// key1: originName key2: isServer
var metricNameDictionary = map[string]map[bool]string{
	constvalues.RequestIo + "_sum":        {true: EntityRequestIoMetric, false: TopologyRequestIoMetric},
	constvalues.ResponseIo + "_sum":       {true: EntityResponseIoMetric, false: TopologyResponseIoMetric},
	constvalues.RequestTotalTime + "_sum": {true: EntityRequestLatencyTotalMetric, false: TopologyRequestLatencyTotalMetric},
	constvalues.RequestCount + "_sum":     {true: EntityRequestCountMetric, false: TopologyRequestCountMetric},
	constvalues.RequestTotalTime + "_avg": {true: EntityRequestLatencyMetric, false: TopologyRequestLatencyMetric},
}

const (
	TopologyRequestIoMetric  = "request_bytes_total"
	TopologyResponseIoMetric = "response_bytes_total"
	// TopologyRequestLatencyMetric is a histogram
	TopologyRequestLatencyMetric      = "duration_nanoseconds"
	TopologyRequestLatencyTotalMetric = "duration_nanoseconds_total"
	TopologyRequestCountMetric        = "total"

	EntityRequestIoMetric  = "receive_bytes_total"
	EntityResponseIoMetric = "send_bytes_total"
	// EntityRequestLatencyMetric is a histogram
	EntityRequestLatencyMetric      = "duration_nanoseconds"
	EntityRequestLatencyTotalMetric = "duration_nanoseconds_total"
	EntityRequestCountMetric        = "total"
)

const (
	NPMPrefixKindling = "kindling"

	EntityPrefix   = "entity"
	TopologyPrefix = "topology"
)

func ToKindlingTraceAsMetricName() string {
	return NPMPrefixKindling + "_trace_request_" + TopologyRequestLatencyMetric
}

func ToKindlingMetricName(origName string, isServer bool) string {
	if names, ok := metricNameDictionary[origName]; !ok {
		return ""
	} else {
		return getKindlingPrefix(isServer) + "request_" + names[isServer]
	}
}

//ToKindlingDetailMetricName For ServerDetail Metric
func ToKindlingDetailMetricName(origName string, protocol string) string {
	if names, ok := metricNameDictionary[origName]; !ok {
		return ""
	} else {
		return getKindlingPrefix(true) + protocol + "_" + names[true]
	}
}

func getKindlingPrefix(isServer bool) string {
	var kindMark string
	if isServer {
		kindMark = EntityPrefix
	} else {
		kindMark = TopologyPrefix
	}
	return NPMPrefixKindling + "_" + kindMark + "_"
}
