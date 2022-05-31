package tcpconnectanalyzer

import (
	"time"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/analyzer/tcpconnectanalyzer/internal"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	conntrackerpackge "github.com/Kindling-project/kindling/collector/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

const Type analyzer.Type = "tcpconnectanalyzer"

var consumableEvents = map[string]bool{
	constnames.ConnectEvent:     true,
	constnames.TcpConnectEvent:  true,
	constnames.TcpSetStateEvent: true,
}

type TcpConnectAnalyzer struct {
	config        *Config
	nextConsumers []consumer.Consumer
	conntracker   conntrackerpackge.Conntracker

	eventChannel   chan *model.KindlingEvent
	connectMonitor *internal.ConnectMonitor

	stopCh chan bool

	telemetry *component.TelemetryTools
}

func New(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer {
	config := cfg.(*Config)
	ret := &TcpConnectAnalyzer{
		config:        config,
		nextConsumers: consumers,
		telemetry:     telemetry,
		eventChannel:  make(chan *model.KindlingEvent, config.ChannelSize),
		stopCh:        make(chan bool),

		connectMonitor: internal.NewConnectMonitor(telemetry.Logger),
	}
	conntracker, err := conntrackerpackge.NewConntracker(nil)
	if err != nil {
		telemetry.Logger.Warn("Conntracker cannot work as expected:", zap.Error(err))
	}
	ret.conntracker = conntracker
	newSelfMetrics(telemetry.MeterProvider, ret.connectMonitor)
	return ret
}

// Start initializes the analyzer
func (a *TcpConnectAnalyzer) Start() error {
	go func() {
		timeoutTicker := time.NewTicker(time.Duration(a.config.TimeoutSecond/5) * time.Second)
		scanTcpStateTicker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-timeoutTicker.C:
				a.trimExpiredConnStats()
			case <-scanTcpStateTicker.C:
				a.trimConnectionsWithTcpStat()
			case event := <-a.eventChannel:
				a.consumeChannelEvent(event)
			case <-a.stopCh:
				// Only trim the connections expired. For those unfinished, we leave them
				// unchanged and just shutdown this goroutine.
				a.trimConnectionsWithTcpStat()
				a.trimExpiredConnStats()
				return
			}
		}
	}()
	return nil
}

// ConsumeEvent gets the event from the previous component
func (a *TcpConnectAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	eventName := event.Name
	if ok := consumableEvents[eventName]; !ok {
		return nil
	}
	a.eventChannel <- event
	return nil
}

func (a *TcpConnectAnalyzer) consumeChannelEvent(event *model.KindlingEvent) {
	var (
		connectStats *internal.ConnectionStats
		err          error
	)

	switch event.Name {
	case constnames.ConnectEvent:
		if !event.IsTcp() {
			return
		}
		connectStats, err = a.connectMonitor.ReadInConnectExitSyscall(event)
	case constnames.TcpConnectEvent:
		connectStats, err = a.connectMonitor.ReadInTcpConnect(event)
	case constnames.TcpSetStateEvent:
		connectStats, err = a.connectMonitor.ReadInTcpSetState(event)
	}

	if err != nil {
		a.telemetry.Logger.Debug("Cannot update connection stats:", zap.Error(err))
		return
	}
	// Connection is not established yet
	if connectStats == nil {
		return
	}

	gaugeGroup := a.generateGaugeGroup(connectStats)
	a.passThroughConsumers(gaugeGroup)
}

func (a *TcpConnectAnalyzer) trimExpiredConnStats() {
	connStats := a.connectMonitor.TrimExpiredConnections(a.config.TimeoutSecond)
	for _, connStat := range connStats {
		gaugeGroup := a.generateGaugeGroup(connStat)
		a.passThroughConsumers(gaugeGroup)
	}
}

func (a *TcpConnectAnalyzer) trimConnectionsWithTcpStat() {
	connStats := a.connectMonitor.TrimConnectionsWithTcpStat()
	for _, connStat := range connStats {
		gaugeGroup := a.generateGaugeGroup(connStat)
		a.passThroughConsumers(gaugeGroup)
	}
}

func (a *TcpConnectAnalyzer) passThroughConsumers(gaugeGroup *model.GaugeGroup) {
	var retError error
	for _, nextConsumer := range a.nextConsumers {
		err := nextConsumer.Consume(gaugeGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	if retError != nil {
		a.telemetry.Logger.Warn("Error happened while passing through processors:", zap.Error(retError))
	}
}

func (a *TcpConnectAnalyzer) generateGaugeGroup(connectStats *internal.ConnectionStats) *model.GaugeGroup {
	labels := model.NewAttributeMap()
	// The connect events always come from the client-side
	labels.AddBoolValue(constlabels.IsServer, false)
	if connectStats.ConnectSyscall != nil {
		labels.AddStringValue(constlabels.ContainerId, connectStats.ConnectSyscall.GetContainerId())
	}
	labels.AddIntValue(constlabels.Errno, int64(connectStats.Code))
	if connectStats.StateMachine.GetCurrentState() == internal.Closed {
		lastState := connectStats.StateMachine.GetLastState()
		if lastState == internal.Success {
			labels.AddBoolValue(constlabels.Success, true)
		} else {
			labels.AddBoolValue(constlabels.Success, false)
		}
	} else if connectStats.StateMachine.GetCurrentState() == internal.Success {
		labels.AddBoolValue(constlabels.Success, true)
	} else {
		labels.AddBoolValue(constlabels.Success, false)
	}
	srcIp := connectStats.ConnKey.SrcIP
	dstIp := connectStats.ConnKey.DstIP
	srcPort := connectStats.ConnKey.SrcPort
	dstPort := connectStats.ConnKey.DstPort
	labels.UpdateAddStringValue(constlabels.SrcIp, srcIp)
	labels.UpdateAddStringValue(constlabels.DstIp, dstIp)
	labels.UpdateAddIntValue(constlabels.SrcPort, int64(srcPort))
	labels.UpdateAddIntValue(constlabels.DstPort, int64(dstPort))
	dNatIp, dNatPort := a.findDNatTuple(srcIp, uint64(srcPort), dstIp, uint64(dstPort))
	labels.AddStringValue(constlabels.DnatIp, dNatIp)
	labels.AddIntValue(constlabels.DnatPort, dNatPort)

	countValue := &model.Gauge{
		Name:  constnames.TcpConnectTotalMetric,
		Value: 1,
	}
	durationValue := &model.Gauge{
		Name:  constnames.TcpConnectDurationMetric,
		Value: connectStats.GetConnectDuration(),
	}

	retGaugeGroup := model.NewGaugeGroup(
		constnames.TcpConnectGaugeGroupName,
		labels,
		connectStats.EndTimestamp,
		countValue, durationValue)

	return retGaugeGroup
}

func (a *TcpConnectAnalyzer) findDNatTuple(sIp string, sPort uint64, dIp string, dPort uint64) (string, int64) {
	dNat := a.conntracker.GetDNATTupleWithString(sIp, dIp, uint16(sPort), uint16(dPort), 0)
	if dNat == nil {
		return "", -1
	}
	dNatIp := dNat.ReplSrcIP.String()
	dNatPort := dNat.ReplSrcPort
	return dNatIp, int64(dNatPort)
}

// Shutdown cleans all the resources used by the analyzer
func (a *TcpConnectAnalyzer) Shutdown() error {
	a.stopCh <- true
	return nil
}

// Type returns the type of the analyzer
func (a *TcpConnectAnalyzer) Type() analyzer.Type {
	return Type
}
