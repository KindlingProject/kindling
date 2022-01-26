package receiver

import (
	"errors"
	analyzerpackage "github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/model"
)

const Mock = "mock_receiver"

type MockReceiver struct {
	cfg             *Config
	analyzerManager analyzerpackage.Manager
	telemetry       *component.TelemetryTools
}

func NewMockReceiver(cfg interface{}, telemetry *component.TelemetryTools, analyzerManager analyzerpackage.Manager) Receiver {
	config, ok := cfg.(*Config)
	if !ok {
		telemetry.Logger.Sugar().Panicf("Cannot convert mock_analyzer config")
	}
	return &MockReceiver{
		cfg:             config,
		analyzerManager: analyzerManager,
		telemetry:       telemetry,
	}
}

func (r *MockReceiver) Start() error {
	r.telemetry.Logger.Sugar().Infof("Start MockReceiver...")
	// Receive events
	events := make([]*model.KindlingEvent, 5)
	// Distribute events to different analyzers
	analyzer, ok := r.analyzerManager.GetAnalyzer(analyzerpackage.MockAnalyzerType)
	if !ok {
		return errors.New("no mock_analyzer found")
	}
	for _, event := range events {
		err := analyzer.ConsumeEvent(event)
		if err != nil {
			r.telemetry.Logger.Sugar().Infof("Failed to consume event: %s, error is: %v", events, err)
			continue
		}
	}
	return nil
}

func (r *MockReceiver) Shutdown() error {

	return nil
}

type Config struct {
	Name string `mapstructure:"name"`
}
