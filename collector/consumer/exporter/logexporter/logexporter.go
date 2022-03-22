package logexporter

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Type = "logexporter"

type LogExporter struct {
	telemetry *component.TelemetryTools
}

func New(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	return &LogExporter{
		telemetry: telemetry,
	}
}

func (e *LogExporter) Consume(gaugeGroup *model.GaugeGroup) error {
	if ce := e.telemetry.Logger.Check(zapcore.DebugLevel, "Receiver GaugeGroup"); ce != nil {
		ce.Write(
			zap.String("gaugeGroup", gaugeGroup.String()),
		)
	}
	return nil
}

type Config struct {
}
