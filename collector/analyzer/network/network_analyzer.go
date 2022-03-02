package network

import (
	"github.com/Kindling-project/kindling/collector/component"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/consumer"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/analyzer/network/protocol/factory"
	"github.com/Kindling-project/kindling/collector/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
)

const (
	CACHE_ADD_THRESHOLD   = 50
	CACHE_RESET_THRESHOLD = 5000

	Network analyzer.Type = "networkanalyzer"
)

type NetworkAnalyzer struct {
	cfg           *Config
	nextConsumers []consumer.Consumer
	conntracker   *conntracker.Conntracker

	staticPortMap    map[uint32]string
	slowThresholdMap map[string]int
	protocolMap      map[string]*protocol.ProtocolParser
	parsers          []*protocol.ProtocolParser

	gaugeGroupPool *GaugeGroupPool
	requestMonitor sync.Map
	telemetry      *component.TelemetryTools
}

func NewNetworkAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	config, _ := cfg.(*Config)
	return &NetworkAnalyzer{
		cfg:            config,
		gaugeGroupPool: NewGaugePool(),
		nextConsumers:  consumers,
		telemetry:      telemetry,
	}
}

func (na *NetworkAnalyzer) Start() error {
	if na.cfg.EnableConntrack {
		na.conntracker, _ = conntracker.NewConntracker(na.cfg.ConntrackMaxStateSize)
	}
	go na.consumerFdNoReusingTrace()

	na.staticPortMap = map[uint32]string{}
	for _, config := range na.cfg.ProtocolConfigs {
		for _, port := range config.Ports {
			na.staticPortMap[port] = config.Key
		}
	}

	na.slowThresholdMap = map[string]int{}
	disableDisernProtocols := map[string]bool{}
	for _, config := range na.cfg.ProtocolConfigs {
		na.slowThresholdMap[config.Key] = config.Threshold
		disableDisernProtocols[config.Key] = config.DisableDiscern
	}

	na.protocolMap = map[string]*protocol.ProtocolParser{}
	parsers := make([]*protocol.ProtocolParser, 0)
	for _, protocol := range na.cfg.ProtocolParser {
		protocolparser := factory.GetParser(protocol)
		if protocolparser != nil {
			na.protocolMap[protocol] = protocolparser
			disableDisern, ok := disableDisernProtocols[protocol]
			if !ok || !disableDisern {
				parsers = append(parsers, protocolparser)
			}
		}
	}
	// Add Generic Last
	parsers = append(parsers, factory.GetGenericParser())
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
		return na.analyseRequest(evt)
	} else {
		return na.analyseResponse(evt)
	}
}

func (na *NetworkAnalyzer) consumerFdNoReusingTrace() {
	timer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timer.C:
			na.requestMonitor.Range(func(k, v interface{}) bool {
				mps := v.(*messagePairs)
				var timeoutTs = mps.getTimeoutTs()
				if timeoutTs != 0 && (time.Now().UnixNano()/1000000000-int64(timeoutTs)/1000000000) >= 15 {
					na.distributeTraceMetric(mps, nil)
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
	}
	return nil
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

		if oldPairs.responses != nil || oldPairs.requests.IsTimeout(evt, na.cfg.GetRequestTimeout()) {
			na.distributeTraceMetric(oldPairs, mps)
		} else {
			oldPairs.mergeRequest(evt)
		}
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

	if newPairs != nil {
		na.requestMonitor.Store(newPairs.getKey(), newPairs)
	} else {
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
		if ce := na.telemetry.Logger.Check(zapcore.DebugLevel, "NetworkAnalyzer To NextProcess: "); ce != nil {
			ce.Write(
				zap.String("record", record.String()),
			)
		}
		for _, nexConsumer := range na.nextConsumers {
			nexConsumer.Consume(record)
		}
		na.gaugeGroupPool.Free(record)
	}
	return nil
}

func (na *NetworkAnalyzer) parseProtocols(mps *messagePairs) []*model.GaugeGroup {
	// Step 1:  Static Config for port and protocol set in config file
	port := mps.getPort()
	staticProtocol, found := na.staticPortMap[port]
	if found {
		if mps.requests == nil {
			// Connect Timeout
			return na.getConnectFailRecords(mps, staticProtocol)
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
		return na.getConnectFailRecords(mps, protocol.GENERIC)
	}

	// Step2 Cache protocol and port
	// TODO There is concurrent modify case when looping. Considering threadsafe.
	cacheParsers, ok := factory.GetCachedParsersByPort(port)
	if ok {
		for _, parser := range cacheParsers {
			records := na.parseProtocol(mps, parser)
			if records != nil {
				if protocol.GENERIC == parser.GetProtocol() {
					// Reset mapping for  generic and port when exceed threshold so as to parsed by other protcols.
					if parser.AddPortCount(port) == CACHE_RESET_THRESHOLD {
						parser.ResetPort(port)
						factory.RemoveCachedParser(port, parser)
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
				factory.AddCachedParser(port, parser)
			}
			return records
		}
	}
	return na.getRecords(mps, protocol.GENERIC, nil)
}

func (na *NetworkAnalyzer) parseProtocol(mps *messagePairs, parser *protocol.ProtocolParser) []*model.GaugeGroup {
	if parser.MultiRequests() {
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

		records := make([]*model.GaugeGroup, 0)
		if mps.responses == nil {
			size := mps.requests.size()
			for i := 0; i < size; i++ {
				req := mps.requests.getEvent(i)
				mp := &messagePair{
					request:  req,
					response: nil,
				}
				records = append(records, na.getRecord(mps, mp, parser.GetProtocol(), parsedReqMsgs[i].GetAttributes()))
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
				records = append(records, na.getRecord(mps, mp, parser.GetProtocol(), responseMsg.GetAttributes()))
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
					records = append(records, na.getRecord(mps, mp, parser.GetProtocol(), parsedReqMsgs[i].GetAttributes()))
				}
			}

			return records
		}
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

func (na *NetworkAnalyzer) getConnectFailRecords(mps *messagePairs, protocol string) []*model.GaugeGroup {
	evt := mps.connects.event
	ret := na.gaugeGroupPool.Get()
	ret.UpdateAddGauge(constvalues.ConnectTime, int64(mps.connects.getDuration()))
	ret.UpdateAddGauge(constvalues.RequestTotalTime, int64(mps.connects.getDuration()))
	ret.Labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
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
	return []*model.GaugeGroup{ret}
}

func (na *NetworkAnalyzer) getRecords(mps *messagePairs, protocol string, attributes *model.AttributeMap) []*model.GaugeGroup {
	evt := mps.requests.event

	slow := false
	if mps.responses != nil {
		slow = na.isSlow(mps.getDuration(), protocol)
	}

	ret := na.gaugeGroupPool.Get()
	labels := ret.Labels
	labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
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
	// If no protocol error found, we check other errors
	if !labels.GetBoolValue(constlabels.IsError) && mps.responses == nil {
		labels.AddBoolValue(constlabels.IsError, true)
		labels.AddIntValue(constlabels.ErrorType, int64(constlabels.NoResponse))
	}

	if nil != mps.natTuple && mps.responses != nil {
		labels.UpdateAddStringValue(constlabels.DnatIp, mps.natTuple.ReplSrcIP.String())
		labels.UpdateAddIntValue(constlabels.DnatPort, int64(mps.natTuple.ReplSrcPort))
	}

	ret.UpdateAddGauge(constvalues.ConnectTime, int64(mps.getConnectDuration()))
	ret.UpdateAddGauge(constvalues.RequestSentTime, mps.getSentTime())
	ret.UpdateAddGauge(constvalues.WaitingTtfbTime, mps.getWaitingTime())
	ret.UpdateAddGauge(constvalues.ContentDownloadTime, mps.getDownloadTime())
	ret.UpdateAddGauge(constvalues.RequestTotalTime, int64(mps.getConnectDuration()+mps.getDuration()))
	ret.UpdateAddGauge(constvalues.RequestIo, int64(mps.getRquestSize()))
	ret.UpdateAddGauge(constvalues.ResponseIo, int64(mps.getResponseSize()))

	ret.Timestamp = evt.GetStartTime()

	return []*model.GaugeGroup{ret}
}

func (na *NetworkAnalyzer) getRecord(mps *messagePairs, mp *messagePair, protocol string, attributes *model.AttributeMap) *model.GaugeGroup {
	evt := mp.request

	slow := na.isSlow(mp.getDuration(), protocol)
	ret := na.gaugeGroupPool.Get()
	labels := ret.Labels
	labels.UpdateAddIntValue(constlabels.Pid, int64(evt.GetPid()))
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
	// If no protocol error found, we check other errors
	if !labels.GetBoolValue(constlabels.IsError) && mps.responses == nil {
		labels.AddBoolValue(constlabels.IsError, true)
		labels.AddIntValue(constlabels.ErrorType, int64(constlabels.NoResponse))
	}

	if nil != mps.natTuple && mps.responses != nil {
		labels.UpdateAddStringValue(constlabels.DnatIp, mps.natTuple.ReplSrcIP.String())
		labels.UpdateAddIntValue(constlabels.DnatPort, int64(mps.natTuple.ReplSrcPort))
	}

	ret.UpdateAddGauge(constvalues.ConnectTime, 0)
	ret.UpdateAddGauge(constvalues.RequestSentTime, mps.getSentTime())
	ret.UpdateAddGauge(constvalues.WaitingTtfbTime, mps.getWaitingTime())
	ret.UpdateAddGauge(constvalues.ContentDownloadTime, mps.getDownloadTime())
	ret.UpdateAddGauge(constvalues.RequestTotalTime, int64(mps.getDuration()))
	ret.UpdateAddGauge(constvalues.RequestIo, int64(mps.getRquestSize()))
	ret.UpdateAddGauge(constvalues.ResponseIo, int64(mps.getResponseSize()))

	ret.Timestamp = evt.GetStartTime()
	return ret
}

func (na *NetworkAnalyzer) isSlow(duration uint64, protocol string) bool {
	return int64(duration) >= int64(na.getResponseSlowThreshold(protocol))*int64(time.Millisecond)
}

func (na *NetworkAnalyzer) getResponseSlowThreshold(protocol string) int {
	if value, ok := na.slowThresholdMap[protocol]; ok {
		return value
	}
	return na.cfg.getResponseSlowThreshold()
}
