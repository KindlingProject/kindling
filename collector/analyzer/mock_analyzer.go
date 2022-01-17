package analyzer

import (
	"github.com/dxsup/kindling-collector/consumer"
	"github.com/dxsup/kindling-collector/model"
	"go.uber.org/zap"
)

const MockAnalyzerType = "mock_analyzer"

type MockAnalyzer struct {
	cfg          *Config
	nextConsumer consumer.Consumer
	logger       *zap.Logger
}

func NewMockAnalyzer(cfg interface{}, logger *zap.Logger, consumer consumer.Consumer) Analyzer {
	config, ok := cfg.(*Config)
	if !ok {
		logger.Panic("Cannot convert mock_analyzer config")
	}
	return &MockAnalyzer{
		cfg:          config,
		nextConsumer: consumer,
		logger:       logger,
	}
}

func (a *MockAnalyzer) Start() error {
	return nil
}

func (a *MockAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	a.logger.Info("[MockAnalyzer] Receive a new event: ", zap.String("event", event.String()))
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
