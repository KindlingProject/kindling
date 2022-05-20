package pgftmetricanalyzer

import (
	"fmt"

	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
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
	consumers []consumer.Consumer
	telemetry *component.TelemetryTools
}

func NewPgftMetricAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, nextConsumers []consumer.Consumer) analyzer.Analyzer {
	retAnalyzer := &PgftMetricAnalyzer{
		consumers: nextConsumers,
		telemetry: telemetry,
	}
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

	pgftMaj := event.GetUserAttribute("pgft_maj")
	pgftMin := event.GetUserAttribute("pgft_min")
	ptMaj := (int64)(pgftMaj.GetUintValue())
	ptMin := (int64)(pgftMin.GetUintValue())

	var gaugeSlice []*model.Gauge
	gaugeMaj := &model.Gauge{
		Name:  constnames.PgftSwitchMajorMetricName,
		Value: ptMaj,
	}

	gaugeMin := &model.Gauge{
		Name:  constnames.PgftSwitchMinorMetricName,
		Value: ptMin,
	}
	if ptMaj != 0 {
		gaugeSlice = append(gaugeSlice, gaugeMaj)
	}
	if ptMin != 0 {
		gaugeSlice = append(gaugeSlice, gaugeMin)
	}

	return model.NewGaugeGroup(constnames.PgftGaugeGroupName, labels, event.Timestamp, gaugeSlice...), nil
}

func (a *PgftMetricAnalyzer) getSwitchLabels(event *model.KindlingEvent) (*model.AttributeMap, error) {

	labels := model.NewAttributeMap()
	ctx := event.GetCtx()
	if ctx == nil {
		return labels, fmt.Errorf("ctx is nil for event %s", event.Name)
	}

	threadinfo := ctx.GetThreadInfo()
	if threadinfo == nil {
		return labels, fmt.Errorf("threadinfo is nil for event %s", event.Name)
	}

	containerId := threadinfo.GetContainerId()
	containerName := threadinfo.GetContainerName()

	tid := (int64)(threadinfo.GetTid())
	pid := (int64)(threadinfo.GetPid())

	labels.AddIntValue(constlabels.Tid, tid)
	labels.AddIntValue(constlabels.Pid, pid)
	labels.AddStringValue(constlabels.ContainerId, containerId)
	labels.AddStringValue(constlabels.Container, containerName)

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
