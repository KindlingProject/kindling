package tcpmetricanalyzer

import (
	"fmt"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	conntrackerpackge "github.com/Kindling-project/kindling/collector/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TcpMetric analyzer.Type = "tcpmetricanalyzer"

	TcpRttMetricName        = "kindling_tcp_rtt_microseconds"
	TcpRetransmitMetricName = "kindling_tcp_retransmit_total"
	TcpDropMetricName       = "kindling_tcp_packet_loss_total"
)

var consumableEvents = map[string]bool{
	constnames.TcpCloseEvent:          true,
	constnames.TcpRcvEstablishedEvent: true,
	constnames.TcpDropEvent:           true,
	constnames.TcpRetransmitSkbEvent:  true,
}

type TcpMetricAnalyzer struct {
	consumers   []consumer.Consumer
	conntracker *conntrackerpackge.Conntracker
	telemetry   *component.TelemetryTools
}

func NewTcpMetricAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, nextConsumers []consumer.Consumer) analyzer.Analyzer {
	retAnalyzer := &TcpMetricAnalyzer{
		consumers: nextConsumers,
		telemetry: telemetry,
	}
	conntracker, err := conntrackerpackge.NewConntracker(10000)
	if err != nil {
		telemetry.Logger.Panic("Failed to create TcpMetricAnalyzer: ", zap.Error(err))
	}
	retAnalyzer.conntracker = conntracker
	return retAnalyzer
}

func (a *TcpMetricAnalyzer) Start() error {
	return nil
}

// ConsumeEvent gets the event from the previous component
func (a *TcpMetricAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	_, ok := consumableEvents[event.Name]
	if !ok {
		return nil
	}
	var gaugeGroup *model.GaugeGroup
	var err error
	switch event.Name {
	case constnames.TcpCloseEvent:
	case constnames.TcpRcvEstablishedEvent:
		gaugeGroup, err = a.generateRtt(event)
	case constnames.TcpDropEvent:
		gaugeGroup, err = a.generateDrop(event)
	case constnames.TcpRetransmitSkbEvent:
		gaugeGroup, err = a.generateRetransmit(event)
	}
	if err != nil {
		if ce := a.telemetry.Logger.Check(zapcore.DebugLevel, "Event Skip, "); ce != nil {
			ce.Write(
				zap.Error(err),
			)
		}
		return nil
	}
	if gaugeGroup == nil {
		return nil
	}
	var retError error
	for _, nextConsumer := range a.consumers {
		err := nextConsumer.Consume(gaugeGroup)
		if err != nil {
			retError = multierror.Append(retError, err)
		}
	}
	return retError
}

func (a *TcpMetricAnalyzer) generateRtt(event *model.KindlingEvent) (*model.GaugeGroup, error) {
	// Only client-side has rtt metric
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}
	// Unit is microsecond
	rtt := event.GetUint64UserAtrribute("rtt")
	// rtt is zero when the kprobe is invoked in the first time, which should be filtered
	if rtt == 0 {
		return nil, nil
	}
	gauge := &model.Gauge{
		Name:  TcpRttMetricName,
		Value: int64(rtt),
	}
	return model.NewGaugeGroup(constnames.TcpGaugeGroupName, labels, event.Timestamp, gauge), nil
}

func (a *TcpMetricAnalyzer) generateRetransmit(event *model.KindlingEvent) (*model.GaugeGroup, error) {
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}
	gauge := &model.Gauge{
		Name:  TcpRetransmitMetricName,
		Value: 1,
	}
	return model.NewGaugeGroup(constnames.TcpGaugeGroupName, labels, event.Timestamp, gauge), nil
}

func (a *TcpMetricAnalyzer) generateDrop(event *model.KindlingEvent) (*model.GaugeGroup, error) {
	labels, err := a.getTupleLabels(event)
	if err != nil {
		return nil, err
	}
	gauge := &model.Gauge{
		Name:  TcpDropMetricName,
		Value: 1,
	}
	return model.NewGaugeGroup(constnames.TcpGaugeGroupName, labels, event.Timestamp, gauge), nil
}

func (a *TcpMetricAnalyzer) getTupleLabels(event *model.KindlingEvent) (*model.AttributeMap, error) {
	// Note: Here sIp/dIp doesn't mean IP from client/server side for sure.
	// sIp stands for the IP which sends tcp flow.
	sIp := event.GetUserAttribute("sip")
	sPort := event.GetUserAttribute("sport")
	dIp := event.GetUserAttribute("dip")
	dPort := event.GetUserAttribute("dport")

	if sIp == nil || sPort == nil || dIp == nil || dPort == nil {
		return nil, fmt.Errorf("one of sip or dip or dport is nil for event %s", event.Name)
	}
	sIpString := model.IPLong2String(uint32(event.GetUint64UserAtrribute("sip")))
	sPortUint := event.GetUint64UserAtrribute("sport")
	dIpString := model.IPLong2String(uint32(event.GetUint64UserAtrribute("dip")))
	dPortUint := event.GetUint64UserAtrribute("dport")

	labels := model.NewAttributeMap()
	labels.AddStringValue(constlabels.SrcIp, sIpString)
	labels.AddIntValue(constlabels.SrcPort, int64(sPortUint))
	labels.AddStringValue(constlabels.DstIp, dIpString)
	labels.AddIntValue(constlabels.DstPort, int64(dPortUint))

	dNatIp, dNatPort := a.findDNatTuple(sIpString, sPortUint, dIpString, dPortUint)
	labels.AddStringValue(constlabels.DnatIp, dNatIp)
	labels.AddIntValue(constlabels.DnatPort, dNatPort)
	return labels, nil
}

func (a *TcpMetricAnalyzer) findDNatTuple(sIp string, sPort uint64, dIp string, dPort uint64) (string, int64) {
	dNat := a.conntracker.GetDNATTupleWithString(sIp, dIp, uint16(sPort), uint16(dPort), 0)
	if dNat == nil {
		return "", -1
	}
	dNatIp := dNat.ReplSrcIP.String()
	dNatPort := dNat.ReplSrcPort
	return dNatIp, int64(dNatPort)
}

// Shutdown cleans all the resources used by the analyzer
func (a *TcpMetricAnalyzer) Shutdown() error {
	return nil
}

// Type returns the type of the analyzer
func (a *TcpMetricAnalyzer) Type() analyzer.Type {
	return TcpMetric
}
