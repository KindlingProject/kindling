package pgftmetricanalyzer

import (
	"log"

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
	PgftMetric analyzer.Type = "pgftmetricanalyzer"
)

var consumableEvents = map[string]bool{
	constnames.SwitchEvent: true,
}

type PgftMetricAnalyzer struct {
	consumers   []consumer.Consumer
	conntracker conntrackerpackge.Conntracker
	telemetry   *component.TelemetryTools
}

func NewPgftMetricAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, nextConsumers []consumer.Consumer) analyzer.Analyzer {
	retAnalyzer := &PgftMetricAnalyzer{
		consumers: nextConsumers,
		telemetry: telemetry,
	}
	conntracker, err := conntrackerpackge.NewConntracker(nil)
	if err != nil {
		telemetry.Logger.Warn("Conntracker cannot work as expected:", zap.Error(err))
	}
	retAnalyzer.conntracker = conntracker
	return retAnalyzer
}

func (a *PgftMetricAnalyzer) Start() error {
	return nil
}

// ConsumeEvent gets the event from the previous component
func (a *PgftMetricAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	_, ok := consumableEvents[event.Name]
	if !ok {
		return nil
	}
	var gaugeGroup *model.GaugeGroup
	var err error
	switch event.Name {
	case constnames.SwitchEvent:
		gaugeGroup, err = a.generateSwitchPgft(event)
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

func (a *PgftMetricAnalyzer) generateSwitchPgft(event *model.KindlingEvent) (*model.GaugeGroup, error) {
	labels, err := a.getSwitchLabels(event)
	if err != nil {
		return nil, err
	}

	gauge := &model.Gauge{
		Name:  constnames.PgftSwitchMetricName,
		Value: 1,
	}
	return model.NewGaugeGroup(constnames.PgftGaugeGroupName, labels, event.Timestamp, gauge), nil
}

func (a *PgftMetricAnalyzer) getSwitchLabels(event *model.KindlingEvent) (*model.AttributeMap, error) {

	next := event.GetUserAttribute("next")
	pgftMaj := event.GetUserAttribute("pgft_maj")
	pgftMin := event.GetUserAttribute("pgft_min")
	vmSize := event.GetUserAttribute("vm_size")
	vmRss := event.GetUserAttribute("vm_rss")
	vmSwap := event.GetUserAttribute("vm_swap")

	ctx := event.GetCtx()
	threadinfo := ctx.GetThreadInfo()
	tid := (int64)(threadinfo.GetTid())
	pid := (int64)(threadinfo.GetPid())
	containerId := threadinfo.GetContainerId()
	containerName := threadinfo.GetContainerName()

	nextPid := (int64)(next.GetIntValue())
	ptMaj := (int64)(pgftMaj.GetUintValue())
	ptMin := (int64)(pgftMin.GetUintValue())
	vmsize := (int64)(vmSize.GetUintValue())
	vmrss := (int64)(vmRss.GetUintValue())
	vmswap := (int64)(vmSwap.GetUintValue())

	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.NextPid, nextPid)
	labels.AddIntValue(constlabels.PgftMaj, ptMaj)
	labels.AddIntValue(constlabels.PgftMin, ptMin)
	labels.AddIntValue(constlabels.VmSize, vmsize)
	labels.AddIntValue(constlabels.VmRss, vmrss)
	labels.AddIntValue(constlabels.VmSwap, vmswap)
	labels.AddIntValue(constlabels.Tid, tid)
	labels.AddIntValue(constlabels.Pid, pid)
	labels.AddStringValue(constlabels.ContainerId, containerId)
	labels.AddStringValue(constlabels.Container, containerName)

	log.Printf("getSwitchLabels\n")
	log.Printf("nextpid: %d, pgftmaj: %d, pgftmin: %d, vmsize: %d, vmrss: %d, vmswap:%d", nextPid, ptMaj, ptMin, vmsize, vmrss, vmswap)

	return labels, nil
}

// Shutdown cleans all the resources used by the analyzer
func (a *PgftMetricAnalyzer) Shutdown() error {
	return nil
}

// Type returns the type of the analyzer
func (a *PgftMetricAnalyzer) Type() analyzer.Type {
	return PgftMetric
}
