package otelexporter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"strconv"
)

type Adapter struct {
	// labelsMap key: protocolType value: a list of realAttributes
	labelsMap  map[extraLabelsKey]realAttributes
	updateKeys []updateKey

	// valueLabelKey dic won't contain origin Key
	valueLabelsFunc valueToLabels

	// to fix some special labels
	adjustLabels []adjustLabels
}

type metricAdapterBuilder struct {
	baseAndCommonLabelsDict []dictionary

	extraLabelsParamList []extraLabelsParam
	extraLabelsKey       []extraLabelsKey
	updateKeys           []updateKey

	// valueLabelKey dic won't contain origin Key
	valueLabelsKey  []dictionary
	valueLabelsFunc valueToLabels

	constLabels []attribute.KeyValue

	adjustLabels []adjustLabels
}

type realAttributes struct {
	// paramMap A list contain baseLabels,commonLabels,extraLabelsParamList,valueLabels,constLabels
	paramMap []attribute.KeyValue
	// metricsDicList A list contain dict of baseLabels,commonLabels,extraLabelsParamList,
	metricsDicList []dictionary
}

type extraLabelsKey struct {
	protocol Protocol
}

type extraLabelsParam struct {
	dicList []dictionary
	extraLabelsKey
}

func updateProtocolKey(key *extraLabelsKey, labels *model.AttributeMap) *extraLabelsKey {
	switch labels.GetStringValue(constlabels.Protocol) {
	case constvalues.ProtocolHttp:
		key.protocol = HTTP
	case constvalues.ProtocolGrpc:
		key.protocol = GRPC
	case constvalues.ProtocolMysql:
		key.protocol = MYSQL
	case constvalues.ProtocolDns:
		key.protocol = DNS
	case constvalues.ProtocolKafka:
		key.protocol = KAFKA
	default:
		key.protocol = UNSUPPORT
	}
	return key
}

type valueToLabels func(gaugeGroup *model.GaugeGroup) []attribute.Value
type updateKey func(key *extraLabelsKey, labels *model.AttributeMap) *extraLabelsKey
type adjustLabels func(labels *model.AttributeMap, attrs []attribute.KeyValue) []attribute.KeyValue

func (key *extraLabelsKey) simpleMergeKey(labelsKey *extraLabelsKey) *extraLabelsKey {
	if key == nil {
		return labelsKey
	}
	if key.protocol == empty {
		key.protocol = labelsKey.protocol
	}
	return key
}

func (param *extraLabelsParam) simpleMergeParam(extraParams *extraLabelsParam) *extraLabelsParam {
	if param == nil {
		return extraParams
	} else {
		param.dicList = append(param.dicList, extraParams.dicList...)
		return param
	}
}

func newAdapterBuilder(
	baseDict []dictionary,
	commonLabels [][]dictionary) *metricAdapterBuilder {

	baseLabels := make([]attribute.KeyValue, len(baseDict))
	// baseLabels
	for j := 0; j < len(baseDict); j++ {
		baseLabels[j].Key = attribute.Key(baseDict[j].newKey)
	}
	// commonLabels
	for j := 0; j < len(commonLabels); j++ {
		for k := 0; k < len(commonLabels[j]); k++ {
			baseLabels = append(baseLabels, attribute.KeyValue{
				Key: attribute.Key(commonLabels[j][k].newKey),
			})
			baseDict = append(baseDict, commonLabels[j][k])
		}
	}

	return &metricAdapterBuilder{
		baseAndCommonLabelsDict: baseDict,
		extraLabelsKey:          make([]extraLabelsKey, 0),
		adjustLabels:            make([]adjustLabels, 0),
	}
}

func (m *metricAdapterBuilder) withExtraLabels(params []extraLabelsParam, update updateKey) *metricAdapterBuilder {
	if m.extraLabelsKey == nil || len(m.extraLabelsKey) == 0 {
		m.extraLabelsKey = make([]extraLabelsKey, len(params))
		for i := 0; i < len(params); i++ {
			m.extraLabelsKey[i] = params[i].extraLabelsKey
		}
		m.extraLabelsParamList = params
		m.updateKeys = make([]updateKey, 1)
		m.updateKeys[0] = update
		return m
	}

	tmpNewExtraParamsList := make([]extraLabelsParam, len(m.extraLabelsParamList)*len(params))
	tmpNewExtraKeyList := make([]extraLabelsKey, len(m.extraLabelsKey)*len(params))

	if len(tmpNewExtraParamsList) != len(tmpNewExtraKeyList) {
		// TODO Error Info!
		return m
	}

	for i := 0; i < len(params); i++ {
		for s := 0; s < len(m.extraLabelsKey); s++ {
			newKey := m.extraLabelsKey[s].simpleMergeKey(&params[i].extraLabelsKey)
			newParam := m.extraLabelsParamList[s].simpleMergeParam(&params[i])

			tmpNewExtraKeyList = append(tmpNewExtraKeyList, *newKey)
			tmpNewExtraParamsList = append(tmpNewExtraParamsList, *newParam)
		}
	}

	m.extraLabelsParamList = tmpNewExtraParamsList
	m.extraLabelsKey = tmpNewExtraKeyList

	return m
}

func (m *metricAdapterBuilder) withValueToLabels(keys []dictionary, valueToLabel valueToLabels) *metricAdapterBuilder {
	m.valueLabelsFunc = valueToLabel
	m.valueLabelsKey = keys
	return m
}

func (m *metricAdapterBuilder) withConstLabels(constLabels []attribute.KeyValue) *metricAdapterBuilder {
	m.constLabels = constLabels
	return m
}

func (m *metricAdapterBuilder) withAdjust(adjustFunc adjustLabels) *metricAdapterBuilder {
	m.adjustLabels = append(m.adjustLabels, adjustFunc)
	return m
}

func (m *metricAdapterBuilder) build() (*Adapter, error) {
	labelsMap := make(map[extraLabelsKey]realAttributes, len(m.extraLabelsKey))
	baseAndCommonParams := make([]attribute.KeyValue, len(m.baseAndCommonLabelsDict))

	for i := 0; i < len(m.baseAndCommonLabelsDict); i++ {
		baseAndCommonParams[i] = attribute.KeyValue{
			Key: attribute.Key(m.baseAndCommonLabelsDict[i].newKey),
		}
	}

	for i := 0; i < len(m.extraLabelsKey); i++ {
		//TODO Check length of extraLabelsKey is equal to extraLabelsParamList , or return error
		//TODO Seem that golang reuse the space of tmpDict in DetailTopologyAdapter unexpected ,need more test
		tmpDict := append(m.baseAndCommonLabelsDict, m.extraLabelsParamList[i].dicList...)
		tmpParamList := baseAndCommonParams
		for s := 0; s < len(m.extraLabelsParamList[i].dicList); s++ {
			tmpParamList = append(tmpParamList, attribute.KeyValue{
				Key: attribute.Key(m.extraLabelsParamList[i].dicList[s].newKey),
			})
		}

		// valueLabels
		if m.valueLabelsKey != nil {
			for s := 0; s < len(m.valueLabelsKey); s++ {
				tmpParamList = append(tmpParamList, attribute.KeyValue{
					Key: attribute.Key(m.valueLabelsKey[s].newKey),
				})
			}
		}

		// constLabels
		if m.constLabels != nil {
			tmpParamList = append(tmpParamList, m.constLabels...)
		}

		labelsMap[m.extraLabelsKey[i]] = realAttributes{
			paramMap:       tmpParamList,
			metricsDicList: tmpDict,
		}
	}

	return &Adapter{
		labelsMap:       labelsMap,
		updateKeys:      m.updateKeys,
		valueLabelsFunc: m.valueLabelsFunc,
		adjustLabels:    m.adjustLabels,
	}, nil
}

func (m *Adapter) adapter(group *model.GaugeGroup) ([]attribute.KeyValue, error) {
	labels := group.Labels
	tmpExtraKey := &extraLabelsKey{protocol: empty}
	for i := 0; i < len(m.updateKeys); i++ {
		tmpExtraKey = m.updateKeys[i](tmpExtraKey, labels)
	}
	attrs := m.labelsMap[*tmpExtraKey]

	for i := 0; i < len(attrs.metricsDicList); i++ {
		switch attrs.metricsDicList[i].valueType {
		case String:
			attrs.paramMap[i].Value = attribute.StringValue(labels.GetStringValue(attrs.metricsDicList[i].originKey))
		case Int64:
			attrs.paramMap[i].Value = attribute.Int64Value(labels.GetIntValue(attrs.metricsDicList[i].originKey))
		case Bool:
			attrs.paramMap[i].Value = attribute.BoolValue(labels.GetBoolValue(attrs.metricsDicList[i].originKey))
		case FromInt64ToString:
			attrs.paramMap[i].Value = attribute.StringValue(strconv.FormatInt(labels.GetIntValue(attrs.metricsDicList[i].originKey), 10))
		case StrEmpty:
			attrs.paramMap[i].Value = attribute.StringValue(constlabels.STR_EMPTY)
		}
	}

	if m.valueLabelsFunc != nil {
		valueLabels := m.valueLabelsFunc(group)
		for i := 0; i < len(valueLabels); i++ {
			attrs.paramMap[i+len(attrs.metricsDicList)].Value = valueLabels[i]
		}
	}

	for i := 0; i < len(m.adjustLabels); i++ {
		attrs.paramMap = m.adjustLabels[i](labels, attrs.paramMap)
	}
	return attrs.paramMap, nil
}
