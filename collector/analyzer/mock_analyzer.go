package analyzer

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
)

const MockAnalyzerType = "mock_analyzer"

type MockAnalyzer struct {
	cfg          *Config
	nextConsumer consumer.Consumer
	telemetry    *component.TelemetryTools
}

func NewMockAnalyzer(cfg interface{}, telemetry *component.TelemetryTools, consumer consumer.Consumer) Analyzer {
	config, ok := cfg.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert mock_analyzer config")
	}
	return &MockAnalyzer{
		cfg:          config,
		nextConsumer: consumer,
		telemetry:    telemetry,
	}
}

func (a *MockAnalyzer) Start() error {
	return nil
}

func (a *MockAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	a.telemetry.Logger.Info("[MockAnalyzer] Receive a new event: ", zap.String("event", event.String()))
	return a.nextConsumer.Consume(&model.GaugeGroup{})
}

func (a *MockAnalyzer) Shutdown() error {
	return nil
}

func (a *MockAnalyzer) Type() Type {
	return MockAnalyzerType
}

type Config struct {
	Num int `yaml:"num"`
}
