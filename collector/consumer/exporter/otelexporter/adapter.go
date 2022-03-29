package otelexporter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type valueType int

type dictionary struct {
	originKey string
	newKey    string
	valueType
}

const (
	Int64  valueType = 0
	String valueType = 1
	Bool   valueType = 3
)

var entityCommonMetricsDicList = []dictionary{
	{"", "", String},
}

var entityAggMetricsDicList = []dictionary{
	{"", "", String},
}

var entityDetailMetricsDicList = []dictionary{
	{"", "", String},
}

var topologyCommonMetricsDicList = []dictionary{
	{"", "", String},
}

var topologyAggMetricsDicList = []dictionary{
	{"", "", String},
}

var topologyDetailMetricsDicList = []dictionary{
	{"", "", String},
}

var ServerEntity *metricAdapter = creatMetricAdapter(nil, entityCommonMetricsDicList, true, nil)

type metricAdapter struct {
	labelsMap            map[extraLabelsKey]realAttributes
	paramMap             []attribute.KeyValue
	metricsDicList       []dictionary
	metricAggregationMap map[string]MetricAggregationKind
	isServer             bool
}

type realAttributes struct {
	paramMap       []attribute.KeyValue
	metricsDicList []dictionary
}

type Protocol int

type Granularity int

const (
	HTTP Protocol = iota
	KAFKA
)

type extraLabelsKey struct {
	protocol Protocol
}

type ProtocolLabels struct {
	Protocol
	labels  []attribute.KeyValue
	dicList []dictionary
}

func withProtocols() []ProtocolLabels {
	return []ProtocolLabels{
		{
			Protocol: HTTP,
			labels:   nil,
			dicList:  nil,
		},
	}
}

func creatMetricAdapter(
	commonLabels []attribute.KeyValue,
	metricsDicList []dictionary,
	isServer bool,
	protocols []ProtocolLabels,
	extraLabels ...[]dictionary) *metricAdapter {
	//TODO Value to labels
	//serviceMetricDefaultLabels := make([]attribute.KeyValue, len(metricsDicList))
	//for i := 0; i < len(metricsDicList); i++ {
	//	serviceMetricDefaultLabels[i].Key = attribute.Key(metricsDicList[i].newKey)
	//}
	//return &metricAdapter{
	//	paramMap:       append(serviceMetricDefaultLabels, commonLabels...),
	//	metricsDicList: metricsDicList,
	//	isServer:       isServer,
	//}

	baseLabels := make([]attribute.KeyValue, len(metricsDicList))
	// baseLabels
	for j := 0; j < len(metricsDicList); j++ {
		baseLabels[j].Key = attribute.Key(metricsDicList[j].newKey)
	}
	// extraLabels
	for j := 0; j < len(extraLabels); j++ {
		for k := 0; k < len(extraLabels[j]); k++ {
			baseLabels = append(baseLabels, attribute.KeyValue{
				Key: attribute.Key(extraLabels[j][k].newKey),
			})
			metricsDicList = append(metricsDicList, extraLabels[j][k])
		}
	}

	realAttributesMap := make(map[extraLabelsKey]realAttributes, len(protocols))
	tmpAttributesKey := extraLabelsKey{}
	for i := 0; i < len(protocols); i++ {
		// Protocols
		tmpAttributesKey.protocol = protocols[i].Protocol
		realAttributesMap[tmpAttributesKey] = realAttributes{
			metricsDicList: append(metricsDicList, protocols[i].dicList...),
		}
	}

	return &metricAdapter{}
}

func (a *metricAdapter) adapter(baseLabels *model.AttributeMap) ([]attribute.KeyValue, error) {
	//TODO protocol labels
	for i := 0; i < len(a.metricsDicList); i++ {
		switch a.metricsDicList[i].valueType {
		case String:
			a.paramMap[i].Value = attribute.StringValue(baseLabels.GetStringValue(a.metricsDicList[i].originKey))
		case Int64:
			a.paramMap[i].Value = attribute.Int64Value(baseLabels.GetIntValue(a.metricsDicList[i].originKey))
		case Bool:
			a.paramMap[i].Value = attribute.BoolValue(baseLabels.GetBoolValue(a.metricsDicList[i].originKey))
		}
	}
	return a.paramMap, nil
}

func (a *metricAdapter) GetMeasurement(insFactory *instrumentFactory, gauges []model.Gauge) {
	// TODO Label to Measurement
	measurements := make([]metric.Measurement, 0, len(gauges))
	for i := 0; i < len(gauges); i++ {
		name := constlabels.ToKindlingMetricName(gauges[i].Name, a.isServer)
		measurements = append(measurements, insFactory.getInstrument(name, a.metricAggregationMap[name]).Measurement(gauges[i].Value))
	}
}
