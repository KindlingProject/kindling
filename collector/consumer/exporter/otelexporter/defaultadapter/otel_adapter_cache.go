package defaultadapter

import (
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"sort"
	"strconv"
)

// adapterCache This struct is an optional component in any adapter.
// Since otlp-sdk is only support to received attribute.values as input labels
// In order to get better performance in some frequent transformations (mainly memory allocation)
// you can refer to this struct to assist in transformation
type adapterCache struct {
	// labelsMap key: protocolType value: a list of realAttributes
	labelsMap  map[extraLabelsKey]realAttributes
	updateKeys []updateKey

	// valueLabelKey dic won't contain origin Key
	valueLabelsFunc valueToLabels

	// to fix some special labels
	adjustFunctions []adjustFunctions
}

type metricAdapterBuilder struct {
	baseAndCommonLabelsDict []dictionary

	extraLabelsParamList []extraLabelsParam
	extraLabelsKey       []extraLabelsKey
	updateKeys           []updateKey

	// valueLabelKey dic won't contain origin Key
	valueLabelsKey  []dictionary
	valueLabelsFunc valueToLabels

	constLabels     []attribute.KeyValue
	adjustFunctions []adjustFunctions
}

type realAttributes struct {
	// paramMap A list contain baseLabels,commonLabels,extraLabelsParamList,valueLabels,constLabels
	paramMap []attribute.KeyValue

	// labelsMap as same as front,just for another output
	AttrsMap *model.AttributeMap

	// metricsDicList A list contain dict of baseLabels,commonLabels,extraLabelsParamList,
	metricsDicList []dictionary
	// sortCache
	sortCache map[int]int
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
type adjustAttrMaps func(labels *model.AttributeMap, attributeMap *model.AttributeMap) *model.AttributeMap
type adjustLabels func(labels *model.AttributeMap, attrs []attribute.KeyValue) []attribute.KeyValue

type adjustFunctions struct {
	adjustAttrMaps adjustAttrMaps
	adjustLabels   adjustLabels
}

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
	if commonLabels != nil {
		for j := 0; j < len(commonLabels); j++ {
			for k := 0; k < len(commonLabels[j]); k++ {
				baseLabels = append(baseLabels, attribute.KeyValue{
					Key: attribute.Key(commonLabels[j][k].newKey),
				})
				baseDict = append(baseDict, commonLabels[j][k])
			}
		}
	}

	return &metricAdapterBuilder{
		baseAndCommonLabelsDict: baseDict,
		extraLabelsKey:          make([]extraLabelsKey, 0),
		adjustFunctions:         make([]adjustFunctions, 0),
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

func (m *metricAdapterBuilder) withAdjust(adjustFunc adjustFunctions) *metricAdapterBuilder {
	m.adjustFunctions = append(m.adjustFunctions, adjustFunc)
	return m
}

func (m *metricAdapterBuilder) build() (*adapterCache, error) {
	labelsMap := make(map[extraLabelsKey]realAttributes, len(m.extraLabelsKey))
	baseAndCommonParams := make([]attribute.KeyValue, len(m.baseAndCommonLabelsDict))

	for i := 0; i < len(m.baseAndCommonLabelsDict); i++ {
		baseAndCommonParams[i] = attribute.KeyValue{
			Key: attribute.Key(m.baseAndCommonLabelsDict[i].newKey),
		}
	}

	for i := 0; i < len(m.extraLabelsKey); i++ {
		//TODO Check length of extraLabelsKey is equal to extraLabelsParamList , or return error
		tmpDict := make([]dictionary, 0, len(m.baseAndCommonLabelsDict)+len(m.extraLabelsParamList[i].dicList))
		tmpDict = append(tmpDict, m.baseAndCommonLabelsDict...)
		tmpDict = append(tmpDict, m.extraLabelsParamList[i].dicList...)
		tmpParamList := make([]attribute.KeyValue, len(baseAndCommonParams))
		copy(tmpParamList, baseAndCommonParams)
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

		// manual sort
		tmpKeysList := make([]string, len(tmpParamList))
		for s := 0; s < len(tmpParamList); s++ {
			tmpKeysList[s] = string(tmpParamList[s].Key)
		}
		sort.Strings(tmpKeysList)
		sortCache := make(map[int]int, len(tmpParamList))
		realParamList := make([]attribute.KeyValue, len(tmpParamList))

		for s := 0; s < len(tmpKeysList); s++ {
			for j := 0; j < len(tmpParamList); j++ {
				if tmpKeysList[s] == string(tmpParamList[j].Key) {
					sortCache[j] = s
					realParamList[s] = tmpParamList[j]
					break
				}
			}
		}

		attrs := make(map[string]model.AttributeValue, len(realParamList))
		attrsMap := model.NewAttributeMapWithValues(attrs)
		for _, label := range m.constLabels {
			switch label.Value.Type() {
			case attribute.STRING:
				attrsMap.AddStringValue(string(label.Key), label.Value.AsString())
			case attribute.INT64:
				attrsMap.AddIntValue(string(label.Key), label.Value.AsInt64())
			case attribute.BOOL:
				attrsMap.AddBoolValue(string(label.Key), label.Value.AsBool())
			}
		}
		labelsMap[m.extraLabelsKey[i]] = realAttributes{
			paramMap:       realParamList,
			AttrsMap:       attrsMap,
			metricsDicList: tmpDict,
			sortCache:      sortCache,
		}
	}

	return &adapterCache{
		labelsMap:       labelsMap,
		updateKeys:      m.updateKeys,
		valueLabelsFunc: m.valueLabelsFunc,
		adjustFunctions: m.adjustFunctions,
	}, nil
}

func (m *adapterCache) transform(group *model.GaugeGroup) (*model.AttributeMap, error) {
	labels := group.Labels
	tmpExtraKey := &extraLabelsKey{protocol: empty}
	for i := 0; i < len(m.updateKeys); i++ {
		tmpExtraKey = m.updateKeys[i](tmpExtraKey, labels)
	}
	attrs := m.labelsMap[*tmpExtraKey]

	for i := 0; i < len(attrs.metricsDicList); i++ {
		switch attrs.metricsDicList[i].valueType {
		case String:
			attrs.AttrsMap.AddStringValue(attrs.metricsDicList[i].newKey, labels.GetStringValue(attrs.metricsDicList[i].originKey))
		case Int64:
			attrs.AttrsMap.AddIntValue(attrs.metricsDicList[i].newKey, labels.GetIntValue(attrs.metricsDicList[i].originKey))
		case Bool:
			attrs.AttrsMap.AddBoolValue(attrs.metricsDicList[i].newKey, labels.GetBoolValue(attrs.metricsDicList[i].originKey))
		case FromInt64ToString:
			attrs.AttrsMap.AddStringValue(attrs.metricsDicList[i].newKey, strconv.FormatInt(labels.GetIntValue(attrs.metricsDicList[i].originKey), 10))
		case StrEmpty:
			attrs.AttrsMap.AddStringValue(attrs.metricsDicList[i].newKey, constlabels.STR_EMPTY)
		}
	}

	if m.valueLabelsFunc != nil {
		valueLabels := m.valueLabelsFunc(group)
		for i := 0; i < len(valueLabels); i++ {
			switch valueLabels[i].Type() {
			case attribute.STRING:
				attrs.AttrsMap.AddStringValue(string(attrs.paramMap[attrs.sortCache[i+len(attrs.metricsDicList)]].Key), valueLabels[i].AsString())
			case attribute.INT64:
				attrs.AttrsMap.AddIntValue(string(attrs.paramMap[attrs.sortCache[i+len(attrs.metricsDicList)]].Key), valueLabels[i].AsInt64())
			case attribute.BOOL:
				attrs.AttrsMap.AddBoolValue(string(attrs.paramMap[attrs.sortCache[i+len(attrs.metricsDicList)]].Key), valueLabels[i].AsBool())
			}
		}
	}

	for i := 0; i < len(m.adjustFunctions); i++ {
		attrs.AttrsMap = m.adjustFunctions[i].adjustAttrMaps(labels, attrs.AttrsMap)
	}
	return attrs.AttrsMap, nil
}

func (m *adapterCache) adapt(group *model.GaugeGroup) ([]attribute.KeyValue, error) {
	labels := group.Labels
	tmpExtraKey := &extraLabelsKey{protocol: empty}
	for i := 0; i < len(m.updateKeys); i++ {
		tmpExtraKey = m.updateKeys[i](tmpExtraKey, labels)
	}
	attrs := m.labelsMap[*tmpExtraKey]

	for i := 0; i < len(attrs.metricsDicList); i++ {
		switch attrs.metricsDicList[i].valueType {
		case String:
			attrs.paramMap[attrs.sortCache[i]].Value = attribute.StringValue(labels.GetStringValue(attrs.metricsDicList[i].originKey))
		case Int64:
			attrs.paramMap[attrs.sortCache[i]].Value = attribute.Int64Value(labels.GetIntValue(attrs.metricsDicList[i].originKey))
		case Bool:
			attrs.paramMap[attrs.sortCache[i]].Value = attribute.BoolValue(labels.GetBoolValue(attrs.metricsDicList[i].originKey))
		case FromInt64ToString:
			attrs.paramMap[attrs.sortCache[i]].Value = attribute.StringValue(strconv.FormatInt(labels.GetIntValue(attrs.metricsDicList[i].originKey), 10))
		case StrEmpty:
			attrs.paramMap[attrs.sortCache[i]].Value = attribute.StringValue(constlabels.STR_EMPTY)
		}
	}

	if m.valueLabelsFunc != nil {
		valueLabels := m.valueLabelsFunc(group)
		for i := 0; i < len(valueLabels); i++ {
			attrs.paramMap[attrs.sortCache[i+len(attrs.metricsDicList)]].Value = valueLabels[i]
		}
	}

	for i := 0; i < len(m.adjustFunctions); i++ {
		attrs.paramMap = m.adjustFunctions[i].adjustLabels(labels, attrs.paramMap)
	}
	return attrs.paramMap, nil
}
