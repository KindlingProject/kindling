package network

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/factory"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"

	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

const (
	CACHE_ADD_THRESHOLD   = 50
	CACHE_RESET_THRESHOLD = 5000

	Network analyzer.Type = "networkanalyzer"
)

type NetworkAnalyzer struct {
	cfg           *Config
	nextConsumers []consumer.Consumer
	conntracker   conntracker.Conntracker

	staticPortMap    map[uint32]string
	slowThresholdMap map[string]int
	protocolMap      map[string]*protocol.ProtocolParser
	parserFactory    *factory.ParserFactory
	parsers          []*protocol.ProtocolParser

	dataGroupPool      DataGroupPool
	requestMonitor     sync.Map
	tcpMessagePairSize int64
	udpMessagePairSize int64
	telemetry          *component.TelemetryTools
}

func NewNetworkAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	config, _ := cfg.(*Config)
	na := &NetworkAnalyzer{
		cfg:           config,
		dataGroupPool: NewDataGroupPool(),
		nextConsumers: consumers,
		telemetry:     telemetry,
	}
	if config.EnableConntrack {
		connConfig := &conntracker.Config{
			Enabled:                      config.EnableConntrack,
			ProcRoot:                     config.ProcRoot,
			ConntrackInitTimeout:         30 * time.Second,
			ConntrackRateLimit:           config.ConntrackRateLimit,
			ConntrackMaxStateSize:        config.ConntrackMaxStateSize,
			EnableConntrackAllNamespaces: true,
		}
		na.conntracker, _ = conntracker.NewConntracker(connConfig)
	}

	na.parserFactory = factory.NewParserFactory(factory.WithUrlClusteringMethod(na.cfg.UrlClusteringMethod))
	return na
}

func (na *NetworkAnalyzer) ConsumableEvents() []string {
	return []string{
		constnames.ReadEvent,
		constnames.WriteEvent,
		constnames.ReadvEvent,
		constnames.WritevEvent,
		constnames.SendToEvent,
		constnames.RecvFromEvent,
		constnames.SendMsgEvent,
		constnames.RecvMsgEvent,
	}
}

func (na *NetworkAnalyzer) Start() error {
	// TODO When import multi annalyzers, this part should move to factory. The metric will relate with analyzers.
	newSelfMetrics(na.telemetry.MeterProvider, na)

	go na.consumerFdNoReusingTrace()
	// go na.consumerUnFinishTrace()
	na.staticPortMap = map[uint32]string{}
	for _, config := range na.cfg.ProtocolConfigs {
		for _, port := range config.Ports {
			na.staticPortMap[port] = config.Key
		}
	}

	na.slowThresholdMap = map[string]int{}
	disableDisernProtocols := map[string]bool{}
	for _, config := range na.cfg.ProtocolConfigs {
		protocol.SetPayLoadLength(config.Key, config.PayloadLength)
		na.slowThresholdMap[config.Key] = config.Threshold
		disableDisernProtocols[config.Key] = config.DisableDiscern
	}

	na.protocolMap = map[string]*protocol.ProtocolParser{}
	parsers := make([]*protocol.ProtocolParser, 0)
	for _, protocol := range na.cfg.ProtocolParser {
		protocolparser := na.parserFactory.GetParser(protocol)
		if protocolparser != nil {
			na.protocolMap[protocol] = protocolparser
			disableDisern, ok := disableDisernProtocols[protocol]
			if !ok || !disableDisern {
				parsers = append(parsers, protocolparser)
			}
		}
	}
	// Add Generic Last
	parsers = append(parsers, na.parserFactory.GetGenericParser())
	na.parsers = parsers

	rand.Seed(time.Now().UnixNano())
	return nil
}

func (na *NetworkAnalyzer) Shutdown() error {
	// TODO: implement
	return nil
}

func (na *NetworkAnalyzer) Type() analyzer.Type {
	return Network
}

func (na *NetworkAnalyzer) ConsumeEvent(evt *model.KindlingEvent) error {
	if evt.Category != model.Category_CAT_NET {
		return nil
	}

	ctx := evt.GetCtx()
	if ctx == nil || ctx.GetThreadInfo() == nil {
		return nil
	}
	fd := ctx.GetFdInfo()
	if fd == nil {
		return nil
	}

	if fd.GetSip() == nil {
		return nil
	}

	// if not dns and udp == 1, return
	if fd.GetProtocol() == model.L4Proto_UDP {
		if _, ok := na.protocolMap[protocol.DNS]; !ok {
			return nil
		}
	}

	if evt.IsConnect() {
		// connect event
		return na.analyseConnect(evt)
	}

	if evt.GetDataLen() <= 0 || evt.GetResVal() < 0 {
		// TODO: analyse udp
		return nil
	}

	isRequest, err := evt.IsRequest()
	if err != nil {
		return err
	}
	if isRequest {
		if evt.GetPid() == 13759 {
			log.Printf("latency = %d", evt.GetLatency())
		}
		return na.analyseRequest(evt)
	} else {
		return na.analyseResponse(evt)
	}
}

func (na *NetworkAnalyzer) consumerFdNoReusingTrace() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			na.requestMonitor.Range(func(k, v interface{}) bool {
				mps := v.(*messagePairs)
				var timeoutTs = mps.getTimeoutTs()
				if timeoutTs != 0 {
					var duration = (time.Now().UnixNano()/1000000000 - int64(timeoutTs)/1000000000)
					if mps.responses != nil && duration >= int64(na.cfg.GetFdReuseTimeout()) {
						// No FdReuse Request
						na.distributeTraceMetric(mps, nil)
					} else if duration >= int64(na.cfg.getNoResponseThreshold()) {
						// No Response Request
						na.distributeTraceMetric(mps, nil)
					}
				}
				return true
			})
		}
	}
}

func (na *NetworkAnalyzer) analyseConnect(evt *model.KindlingEvent) error {
	mps := &messagePairs{
		connects:  newEvents(evt),
		requests:  nil,
		responses: nil,
		mutex:     sync.RWMutex{},
	}
	if pairInterface, exist := na.requestMonitor.LoadOrStore(mps.getKey(), mps); exist {
		// There is an old message pair
		var oldPairs = pairInterface.(*messagePairs)
		// TODO: is there any need to check old connect event?
		if oldPairs.requests == nil && oldPairs.connects != nil {
			if oldPairs.connects.IsTimeout(evt, na.cfg.GetConnectTimeout()) {
				na.distributeTraceMetric(oldPairs, mps)
			} else {
				oldPairs.mergeConnect(evt)
			}
			return nil
		}

		na.distributeTraceMetric(oldPairs, mps)
	} else {
		na.recordMessagePairSize(evt, 1)
	}
	return nil
}

func (na *NetworkAnalyzer) recordMessagePairSize(evt *model.KindlingEvent, count int64) {
	if evt.IsUdp() == 1 {
		atomic.AddInt64(&na.udpMessagePairSize, count)
	} else {
		atomic.AddInt64(&na.tcpMessagePairSize, count)
	}
}

func (na *NetworkAnalyzer) analyseRequest(evt *model.KindlingEvent) error {
	mps := &messagePairs{
		connects:  nil,
		requests:  newEvents(evt),
		responses: nil,
		mutex:     sync.RWMutex{}}
	if pairInterface, exist := na.requestMonitor.LoadOrStore(mps.getKey(), mps); exist {
		// There is an old message pair
		var oldPairs = pairInterface.(*messagePairs)
		if oldPairs.requests == nil {
			if oldPairs.connects == nil {
				// empty message pair, store new one
				na.requestMonitor.Store(mps.getKey(), mps)
				return nil
			} else {
				// there is a connect event, update it
				oldPairs.mergeRequest(evt)
				na.requestMonitor.Store(oldPairs.getKey(), oldPairs)
				return nil
			}
		}

		if oldPairs.responses != nil || oldPairs.requests.IsSportChanged(evt) {
			na.distributeTraceMetric(oldPairs, mps)
		} else {
			oldPairs.mergeRequest(evt)
		}
	} else {
		na.recordMessagePairSize(evt, 1)
	}
	return nil
}

func (na *NetworkAnalyzer) analyseResponse(evt *model.KindlingEvent) error {
	pairInterface, ok := na.requestMonitor.Load(getMessagePairKey(evt))
	if !ok {
		return nil
	}
	var oldPairs = pairInterface.(*messagePairs)
	if oldPairs.requests == nil {
		// empty request, not a valid state
		return nil
	}

	oldPairs.mergeResponse(evt)
	na.requestMonitor.Store(oldPairs.getKey(), oldPairs)
	return nil
}

func (na *NetworkAnalyzer) distributeTraceMetric(oldPairs *messagePairs, newPairs *messagePairs) error {
	var queryEvt *model.KindlingEvent
	if oldPairs.connects != nil {
		queryEvt = oldPairs.connects.event
	} else if oldPairs.requests != nil {
		queryEvt = oldPairs.requests.event
	} else {
		return nil
	}

	if oldPairs.checkSend() == false {
		// FIX send twice for request/response with 15s delay.
		return nil
	}

	if newPairs != nil {
		na.requestMonitor.Store(newPairs.getKey(), newPairs)
	} else {
		na.recordMessagePairSize(queryEvt, -1)
		na.requestMonitor.Delete(oldPairs.getKey())
	}

	// Relate conntrack
	if na.cfg.EnableConntrack {
		srcIP := queryEvt.GetCtx().FdInfo.Sip[0]
		dstIP := queryEvt.GetCtx().FdInfo.Dip[0]
		srcPort := uint16(queryEvt.GetSport())
		dstPort := uint16(queryEvt.GetDport())
		isUdp := queryEvt.IsUdp()
		natTuple := na.conntracker.GetDNATTuple(srcIP, dstIP, srcPort, dstPort, isUdp)
		if nil != natTuple {
			oldPairs.natTuple = natTuple
		}
	}

	// Parse Protocols
	// Case 1 ConnectFail    Connect
	// Case 2 Request 498   Connect/Request                         Request
	// Case 3 Normal             Connect/Request/Response   Request/Response
	records := na.parseProtocols(oldPairs)
	for _, record := range records {
		if ce := na.telemetry.Logger.Check(zapcore.DebugLevel, ""); ce != nil {
			na.telemetry.Logger.Debug("NetworkAnalyzer To NextProcess:\n" + record.String())
		}
		netanalyzerParsedRequestTotal.Add(context.Background(), 1, attribute.String("protocol", record.Labels.GetStringValue(constlabels.Protocol)))
		for _, nexConsumer := range na.nextConsumers {
			nexConsumer.Consume(record)
		}
		na.dataGroupPool.Free(record)
	}
	return nil
}

func (na *NetworkAnalyzer) parseProtocols(mps *messagePairs) []*model.DataGroup {
	// Step 1:  Static Config for port and protocol set in config file
	port := mps.getPort()
	staticProtocol, found := na.staticPortMap[port]
	if found {
		if mps.requests == nil {
			// Connect Timeout
			return na.getConnectFailRecords(mps)
		}

		if parser, exist := na.protocolMap[staticProtocol]; exist {
			records := na.parseProtocol(mps, parser)
			if records != nil {
				return records
			}
		}
		// Return Protocol Only
		// 1. Parser is not implemnet or not set
		// 2. Parse failure
		return na.getRecords(mps, staticProtocol, nil)
	}

	if mps.requests == nil {
		// Connect Timeout
		return na.getConnectFailRecords(mps)
	}

	// Step2 Cache protocol and port
	// TODO There is concurrent modify case when looping. Considering threadsafe.
	cacheParsers, ok := na.parserFactory.GetCachedParsersByPort(port)
	if ok {
		for _, parser := range cacheParsers {
			records := na.parseProtocol(mps, parser)
			if records != nil {
				if protocol.NOSUPPORT == parser.GetProtocol() {
					// Reset mapping for  generic and port when exceed threshold so as to parsed by other protcols.
					if parser.AddPortCount(port) == CACHE_RESET_THRESHOLD {
						parser.ResetPort(port)
						na.parserFactory.RemoveCachedParser(port, parser)
					}
				}
				return records
			}
		}
	}

	// Step3 Loop all protocols
	for _, parser := range na.parsers {
		records := na.parseProtocol(mps, parser)
		if records != nil {
			// Add mapping for port and protocol when exceed threshold
			if parser.AddPortCount(port) == CACHE_ADD_THRESHOLD {
				na.parserFactory.AddCachedParser(port, parser)
			}
			return records
		}
	}
	return na.getRecords(mps, protocol.NOSUPPORT, nil)
}

func (na *NetworkAnalyzer) parseProtocol(mps *messagePairs, parser *protocol.ProtocolParser) []*model.DataGroup {
	if parser.MultiRequests() {
		// Not mergable requests
		return na.parseMultipleRequests(mps, parser)
	}

	// Mergable Data
	requestMsg := protocol.NewRequestMessage(mps.requests.getData())
	if !parser.ParseRequest(requestMsg) {
		// Parse failure
		return nil
	}

	if mps.responses == nil {
		return na.getRecords(mps, parser.GetProtocol(), requestMsg.GetAttributes())
	}

	responseMsg := protocol.NewResponseMessage(mps.responses.getData(), requestMsg.GetAttributes())
	if !parser.ParseResponse(responseMsg) {
		// Parse failure
		return nil
	}
	return na.getRecords(mps, parser.GetProtocol(), responseMsg.GetAttributes())
}

// parseMultipleRequests parses the messagePairs when we know there could be multiple read requests.
// This is used only when the protocol is DNS now.
func (na *NetworkAnalyzer) parseMultipleRequests(mps *messagePairs, parser *protocol.ProtocolParser) []*model.DataGroup {
	// Match with key when disordering.
	size := mps.requests.size()
	parsedReqMsgs := make([]*protocol.PayloadMessage, size)
	for i := 0; i < size; i++ {
		req := mps.requests.getEvent(i)
		requestMsg := protocol.NewRequestMessage(req.GetData())
		if !parser.ParseRequest(requestMsg) {
			// Parse failure
			return nil
		}
		parsedReqMsgs[i] = requestMsg
	}

	records := make([]*model.DataGroup, 0)
	if mps.responses == nil {
		size := mps.requests.size()
		for i := 0; i < size; i++ {
			req := mps.requests.getEvent(i)
			mp := &messagePair{
				request:  req,
				response: nil,
			}
			records = append(records, na.getRecordWithSinglePair(mps, mp, parser.GetProtocol(), parsedReqMsgs[i].GetAttributes()))
		}
		return records
	} else {
		matchedRequestIdx := make(map[int]bool)
		size := mps.responses.size()
		for i := 0; i < size; i++ {
			resp := mps.responses.getEvent(i)
			responseMsg := protocol.NewResponseMessage(resp.GetData(), model.NewAttributeMap())
			if !parser.ParseResponse(responseMsg) {
				// Parse failure
				return nil
			}
			// Match Request with repsone
			matchIdx := parser.PairMatch(parsedReqMsgs, responseMsg)
			if matchIdx == -1 {
				return nil
			}
			matchedRequestIdx[matchIdx] = true

			mp := &messagePair{
				request:  mps.requests.getEvent(matchIdx),
				response: resp,
			}
			records = append(records, na.getRecordWithSinglePair(mps, mp, parser.GetProtocol(), responseMsg.GetAttributes()))
		}
		// 498 Case
		reqSize := mps.requests.size()
		for i := 0; i < reqSize; i++ {
			req := mps.requests.getEvent(i)
			if _, matched := matchedRequestIdx[i]; !matched {
				mp := &messagePair{
					request:  req,
					response: nil,
				}
				records = append(records, na.getRecordWithSinglePair(mps, mp, parser.GetProtocol(), parsedReqMsgs[i].GetAttributes()))
			}
		}
		return records
	}
}

func (na *NetworkAnalyzer) getConnectFailRecords(mps *messagePairs) []*model.DataGroup {
	evt := mps.connects.event
	ret := na.dataGroupPool.Get()
	ret.UpdateAddIntMetric(constvalues.ConnectTime, int64(mps.connects.getDuration()))
	ret.UpdateAddIntMetric(constvalues.RequestTotalTime, int64(mps.connects.getDuration()))
	ret.Labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
	ret.Labels.UpdateAddIntValue(constlabels.RequestTid, 0)
	ret.Labels.UpdateAddIntValue(constlabels.ResponseTid, 0)
	ret.Labels.UpdateAddStringValue(constlabels.Comm, evt.GetComm())
	ret.Labels.UpdateAddStringValue(constlabels.SrcIp, evt.GetSip())
	ret.Labels.UpdateAddStringValue(constlabels.DstIp, evt.GetDip())
	ret.Labels.UpdateAddIntValue(constlabels.SrcPort, int64(evt.GetSport()))
	ret.Labels.UpdateAddIntValue(constlabels.DstPort, int64(evt.GetDport()))
	ret.Labels.UpdateAddStringValue(constlabels.DnatIp, constlabels.STR_EMPTY)
	ret.Labels.UpdateAddIntValue(constlabels.DnatPort, -1)
	ret.Labels.UpdateAddStringValue(constlabels.ContainerId, evt.GetContainerId())
	ret.Labels.UpdateAddBoolValue(constlabels.IsError, true)
	ret.Labels.UpdateAddIntValue(constlabels.ErrorType, int64(constlabels.ConnectFail))
	ret.Labels.UpdateAddBoolValue(constlabels.IsSlow, false)
	ret.Labels.UpdateAddBoolValue(constlabels.IsServer, evt.GetCtx().GetFdInfo().Role)
	ret.Timestamp = evt.GetStartTime()
	return []*model.DataGroup{ret}
}

func (na *NetworkAnalyzer) getRecords(mps *messagePairs, protocol string, attributes *model.AttributeMap) []*model.DataGroup {
	evt := mps.requests.event
	// See the issue https://github.com/KindlingProject/kindling/issues/388 for details.
	if attributes.HasAttribute(constlabels.HttpContinue) {
		if pairInterface, ok := na.requestMonitor.Load(getMessagePairKey(evt)); ok {
			var oldPairs = pairInterface.(*messagePairs)
			oldPairs.putRequestBack(mps.requests)
		}
		return []*model.DataGroup{}
	}

	slow := false
	if mps.responses != nil {
		slow = na.isSlow(mps.getDuration(), protocol)
	}

	ret := na.dataGroupPool.Get()
	labels := ret.Labels
	labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
	addTid(labels, evt, mps.responses)
	labels.UpdateAddStringValue(constlabels.Comm, evt.GetComm())
	labels.UpdateAddStringValue(constlabels.SrcIp, evt.GetSip())
	labels.UpdateAddStringValue(constlabels.DstIp, evt.GetDip())
	labels.UpdateAddIntValue(constlabels.SrcPort, int64(evt.GetSport()))
	labels.UpdateAddIntValue(constlabels.DstPort, int64(evt.GetDport()))
	labels.UpdateAddStringValue(constlabels.DnatIp, constlabels.STR_EMPTY)
	labels.UpdateAddIntValue(constlabels.DnatPort, -1)
	labels.UpdateAddStringValue(constlabels.ContainerId, evt.GetContainerId())
	labels.UpdateAddBoolValue(constlabels.IsError, false)
	labels.UpdateAddIntValue(constlabels.ErrorType, int64(constlabels.NoError))
	labels.UpdateAddBoolValue(constlabels.IsSlow, slow)
	labels.UpdateAddBoolValue(constlabels.IsServer, evt.GetCtx().GetFdInfo().Role)
	labels.UpdateAddStringValue(constlabels.Protocol, protocol)

	labels.Merge(attributes)

	if mps.responses != nil {
		endTimestamp := mps.responses.getLastTimestamp()
		labels.UpdateAddIntValue(constlabels.EndTimestamp, int64(endTimestamp))
	}

	if mps.responses == nil {
		addProtocolPayload(protocol, labels, mps.requests.getData(), nil)
	} else {
		addProtocolPayload(protocol, labels, mps.requests.getData(), mps.responses.getData())
	}

	// If no protocol error found, we check other errors
	if !labels.GetBoolValue(constlabels.IsError) && mps.responses == nil {
		labels.AddBoolValue(constlabels.IsError, true)
		labels.AddIntValue(constlabels.ErrorType, int64(constlabels.NoResponse))
	}

	if nil != mps.natTuple {
		labels.UpdateAddStringValue(constlabels.DnatIp, mps.natTuple.ReplSrcIP.String())
		labels.UpdateAddIntValue(constlabels.DnatPort, int64(mps.natTuple.ReplSrcPort))
	}

	ret.UpdateAddIntMetric(constvalues.ConnectTime, int64(mps.getConnectDuration()))
	ret.UpdateAddIntMetric(constvalues.RequestSentTime, mps.getSentTime())
	ret.UpdateAddIntMetric(constvalues.WaitingTtfbTime, mps.getWaitingTime())
	ret.UpdateAddIntMetric(constvalues.ContentDownloadTime, mps.getDownloadTime())
	ret.UpdateAddIntMetric(constvalues.RequestTotalTime, int64(mps.getConnectDuration()+mps.getDuration()))
	ret.UpdateAddIntMetric(constvalues.RequestIo, int64(mps.getRquestSize()))
	ret.UpdateAddIntMetric(constvalues.ResponseIo, int64(mps.getResponseSize()))

	ret.Timestamp = evt.GetStartTime()

	return []*model.DataGroup{ret}
}

// getRecordWithSinglePair generates a record whose metrics are copied from the input messagePair,
// instead of messagePairs. This is used only when there could be multiple real requests in messagePairs.
// For now, only messagePairs with DNS protocol would run into this method.
func (na *NetworkAnalyzer) getRecordWithSinglePair(mps *messagePairs, mp *messagePair, protocol string, attributes *model.AttributeMap) *model.DataGroup {
	evt := mp.request

	slow := na.isSlow(mp.getDuration(), protocol)
	ret := na.dataGroupPool.Get()
	labels := ret.Labels
	labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
	addTid(labels, evt, mps.responses)
	labels.UpdateAddStringValue(constlabels.Comm, evt.GetComm())
	labels.UpdateAddStringValue(constlabels.SrcIp, evt.GetSip())
	labels.UpdateAddStringValue(constlabels.DstIp, evt.GetDip())
	labels.UpdateAddIntValue(constlabels.SrcPort, int64(evt.GetSport()))
	labels.UpdateAddIntValue(constlabels.DstPort, int64(evt.GetDport()))
	labels.UpdateAddStringValue(constlabels.DnatIp, constlabels.STR_EMPTY)
	labels.UpdateAddIntValue(constlabels.DnatPort, -1)
	labels.UpdateAddStringValue(constlabels.ContainerId, evt.GetContainerId())
	labels.UpdateAddBoolValue(constlabels.IsError, false)
	labels.UpdateAddIntValue(constlabels.ErrorType, int64(constlabels.NoError))
	labels.UpdateAddBoolValue(constlabels.IsSlow, slow)
	labels.UpdateAddBoolValue(constlabels.IsServer, evt.GetCtx().GetFdInfo().Role)
	labels.UpdateAddStringValue(constlabels.Protocol, protocol)

	labels.Merge(attributes)
	if mps.responses != nil {
		endTimestamp := mp.response.Timestamp
		labels.UpdateAddIntValue(constlabels.EndTimestamp, int64(endTimestamp))
	}
	if mp.response == nil {
		addProtocolPayload(protocol, labels, evt.GetData(), nil)
	} else {
		addProtocolPayload(protocol, labels, evt.GetData(), mp.response.GetData())
	}

	// If no protocol error found, we check other errors
	if !labels.GetBoolValue(constlabels.IsError) && mps.responses == nil {
		labels.AddBoolValue(constlabels.IsError, true)
		labels.AddIntValue(constlabels.ErrorType, int64(constlabels.NoResponse))
	}

	if nil != mps.natTuple {
		labels.UpdateAddStringValue(constlabels.DnatIp, mps.natTuple.ReplSrcIP.String())
		labels.UpdateAddIntValue(constlabels.DnatPort, int64(mps.natTuple.ReplSrcPort))
	}

	ret.UpdateAddIntMetric(constvalues.ConnectTime, 0)
	ret.UpdateAddIntMetric(constvalues.RequestSentTime, mp.getSentTime())
	ret.UpdateAddIntMetric(constvalues.WaitingTtfbTime, mp.getWaitingTime())
	ret.UpdateAddIntMetric(constvalues.ContentDownloadTime, mp.getDownloadTime())
	ret.UpdateAddIntMetric(constvalues.RequestTotalTime, int64(mp.getDuration()))
	ret.UpdateAddIntMetric(constvalues.RequestIo, int64(mp.getRquestSize()))
	ret.UpdateAddIntMetric(constvalues.ResponseIo, int64(mp.getResponseSize()))

	ret.Timestamp = evt.GetStartTime()
	return ret
}

func addTid(labels *model.AttributeMap, evt *model.KindlingEvent, responses *events) {
	labels.UpdateAddIntValue(constlabels.RequestTid, int64(evt.GetTid()))
	if responses != nil {
		labels.UpdateAddIntValue(constlabels.ResponseTid, int64(responses.event.GetTid()))
	} else {
		labels.UpdateAddIntValue(constlabels.ResponseTid, 0)
	}
}

func addProtocolPayload(protocolName string, labels *model.AttributeMap, request []byte, response []byte) {
	labels.UpdateAddStringValue(constlabels.RequestPayload, protocol.GetPayloadString(request, protocolName))
	if response != nil {
		labels.UpdateAddStringValue(constlabels.ResponsePayload, protocol.GetPayloadString(response, protocolName))
	} else {
		labels.UpdateAddStringValue(constlabels.ResponsePayload, "")
	}
}

func (na *NetworkAnalyzer) isSlow(duration uint64, protocol string) bool {
	return int64(duration) >= int64(na.getResponseSlowThreshold(protocol))*int64(time.Millisecond)
}

func (na *NetworkAnalyzer) getResponseSlowThreshold(protocol string) int {
	if value, ok := na.slowThresholdMap[protocol]; ok && value > 0 {
		// If value is not set, use response_slow_threshold by default.
		return value
	}
	return na.cfg.getResponseSlowThreshold()
}
