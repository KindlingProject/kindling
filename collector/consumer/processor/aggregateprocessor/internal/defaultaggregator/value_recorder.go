package defaultaggregator

import (
	"github.com/Kindling-project/kindling/collector/consumer/processor/aggregateprocessor/internal"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"sync"
)

type valueRecorder struct {
	name        string
	t           timer
	labelValues map[labelsKey]aggValuesMap
	// mutex is only used to make sure the access to the labelValues is thread-safe.
	// aggValuesMap is responsible for its own thread-safe access.
	mutex      sync.RWMutex
	aggKindMap map[string]aggregatorKind
}

func newValueRecorder(recorderName string, firstDataTimestamp uint64, aggKindMap map[string]aggregatorKind) *valueRecorder {
	return &valueRecorder{
		name:        recorderName,
		t:           newTimer(firstDataTimestamp),
		labelValues: make(map[labelsKey]aggValuesMap),
		mutex:       sync.RWMutex{},
		aggKindMap:  aggKindMap,
	}
}

// Record is thread-safe, and return the result value
func (r *valueRecorder) Record(key *labelsKey, gaugeValues []*model.Gauge) {
	if key == nil {
		return
	}
	r.mutex.RLock()
	aggValues, ok := r.labelValues[*key]
	r.mutex.RUnlock()
	if !ok {
		r.mutex.Lock()
		// double check to avoid double writing
		aggValues, ok = r.labelValues[*key]
		if !ok {
			aggValues = newAggValuesMap(r.aggKindMap)
			r.labelValues[*key] = aggValues
		}
		r.mutex.Unlock()
	}
	for _, gauge := range gaugeValues {
		aggValues.calculate(gauge.Name, gauge.Value)
	}
}

// dump a set of metric from counter cache.
// The return value holds the reference to the metric, not the copied one.
func (r *valueRecorder) dump() []*model.GaugeGroup {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	ret := make([]*model.GaugeGroup, len(r.labelValues))
	index := 0
	for k, v := range r.labelValues {
		gaugeGroup := model.NewGaugeGroup(r.name, k.toLabels(), r.t.outputTimestamp(), v.getAll()...)
		ret[index] = gaugeGroup
		index++
	}
	return ret
}

func (r *valueRecorder) reset() {
	r.mutex.Lock()
	r.labelValues = make(map[labelsKey]aggValuesMap)
	r.mutex.Unlock()
}

type labelsKey struct {
	pid         int64
	protocol    string
	isServer    bool
	containerId string

	srcNode         string
	srcNodeIp       string
	srcNamespace    string
	srcPod          string
	srcWorkloadName string
	srcWorkloadKind string
	srcService      string
	srcIp           string
	srcContainerId  string
	srcContainer    string

	dstNode         string
	dstNodeIp       string
	dstNamespace    string
	dstPod          string
	dstWorkloadName string
	dstWorkloadKind string
	dstService      string
	dstIp           string
	dstPort         int64
	dnatIp          string
	dnatPort        int64
	dstContainerId  string
	dstContainer    string

	httpStatusCode int64
	DnsRcode       int64
	SqlErrCode     int64

	contentKey string
	dnsDomain  string
	kafkaTopic string
}

// TODO: Implement filters
func newLabelsKey(labels *model.AttributeMap, _ *internal.LabelFilter) *labelsKey {
	ret := &labelsKey{
		pid:             labels.GetIntValue(constlabels.Pid),
		protocol:        labels.GetStringValue(constlabels.Protocol),
		isServer:        labels.GetBoolValue(constlabels.IsServer),
		containerId:     labels.GetStringValue(constlabels.ContainerId),
		srcNode:         labels.GetStringValue(constlabels.SrcNode),
		srcNodeIp:       labels.GetStringValue(constlabels.SrcNodeIp),
		srcNamespace:    labels.GetStringValue(constlabels.SrcNamespace),
		srcPod:          labels.GetStringValue(constlabels.SrcPod),
		srcWorkloadName: labels.GetStringValue(constlabels.SrcWorkloadName),
		srcWorkloadKind: labels.GetStringValue(constlabels.SrcWorkloadKind),
		srcService:      labels.GetStringValue(constlabels.SrcService),
		srcIp:           labels.GetStringValue(constlabels.SrcIp),
		srcContainerId:  labels.GetStringValue(constlabels.SrcContainerId),
		srcContainer:    labels.GetStringValue(constlabels.SrcContainer),
		dstNode:         labels.GetStringValue(constlabels.DstNode),
		dstNodeIp:       labels.GetStringValue(constlabels.DstNodeIp),
		dstNamespace:    labels.GetStringValue(constlabels.DstNamespace),
		dstPod:          labels.GetStringValue(constlabels.DstPod),
		dstWorkloadName: labels.GetStringValue(constlabels.DstWorkloadName),
		dstWorkloadKind: labels.GetStringValue(constlabels.DstWorkloadKind),
		dstService:      labels.GetStringValue(constlabels.DstService),
		dstIp:           labels.GetStringValue(constlabels.DstIp),
		dstPort:         labels.GetIntValue(constlabels.DstPort),
		dnatIp:          labels.GetStringValue(constlabels.DnatIp),
		dnatPort:        labels.GetIntValue(constlabels.DnatPort),
		dstContainerId:  labels.GetStringValue(constlabels.DstContainerId),
		dstContainer:    labels.GetStringValue(constlabels.DstContainer),
		httpStatusCode:  labels.GetIntValue(constlabels.HttpStatusCode),
		DnsRcode:        labels.GetIntValue(constlabels.DnsRcode),
		SqlErrCode:      labels.GetIntValue(constlabels.SqlErrCode),
		contentKey:      labels.GetStringValue(constlabels.ContentKey),
		dnsDomain:       labels.GetStringValue(constlabels.DnsDomain),
		kafkaTopic:      labels.GetStringValue(constlabels.KafkaTopic),
	}
	return ret
}

func (k *labelsKey) toLabels() *model.AttributeMap {
	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.Pid, k.pid)
	labels.AddStringValue(constlabels.Protocol, k.protocol)
	labels.AddBoolValue(constlabels.IsServer, k.isServer)
	labels.AddStringValue(constlabels.ContainerId, k.containerId)
	labels.AddStringValue(constlabels.SrcNode, k.srcNode)
	labels.AddStringValue(constlabels.SrcNodeIp, k.srcNodeIp)
	labels.AddStringValue(constlabels.SrcNamespace, k.srcNamespace)
	labels.AddStringValue(constlabels.SrcPod, k.srcPod)
	labels.AddStringValue(constlabels.SrcWorkloadName, k.srcWorkloadName)
	labels.AddStringValue(constlabels.SrcWorkloadKind, k.srcWorkloadKind)
	labels.AddStringValue(constlabels.SrcService, k.srcService)
	labels.AddStringValue(constlabels.SrcIp, k.srcIp)
	labels.AddStringValue(constlabels.SrcContainerId, k.srcContainerId)
	labels.AddStringValue(constlabels.SrcContainer, k.srcContainer)
	labels.AddStringValue(constlabels.DstNode, k.dstNode)
	labels.AddStringValue(constlabels.DstNodeIp, k.dstNodeIp)
	labels.AddStringValue(constlabels.DstNamespace, k.dstNamespace)
	labels.AddStringValue(constlabels.DstPod, k.dstPod)
	labels.AddStringValue(constlabels.DstWorkloadName, k.dstWorkloadName)
	labels.AddStringValue(constlabels.DstWorkloadKind, k.dstWorkloadKind)
	labels.AddStringValue(constlabels.DstService, k.dstService)
	labels.AddStringValue(constlabels.DstIp, k.dstIp)
	labels.AddIntValue(constlabels.DstPort, k.dstPort)
	labels.AddStringValue(constlabels.DnatIp, k.dnatIp)
	labels.AddIntValue(constlabels.DnatPort, k.dnatPort)
	labels.AddStringValue(constlabels.DstContainerId, k.dstContainerId)
	labels.AddStringValue(constlabels.DstContainer, k.dstContainer)
	labels.AddIntValue(constlabels.HttpStatusCode, k.httpStatusCode)
	labels.AddIntValue(constlabels.DnsRcode, k.DnsRcode)
	labels.AddIntValue(constlabels.SqlErrCode, k.SqlErrCode)
	labels.AddStringValue(constlabels.ContentKey, k.contentKey)
	labels.AddStringValue(constlabels.DnsDomain, k.dnsDomain)
	labels.AddStringValue(constlabels.KafkaTopic, k.kafkaTopic)
	return labels
}
