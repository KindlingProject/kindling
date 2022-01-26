package receiver

import (
	"errors"
	analyzerpackage "github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/model"
	"go.uber.org/zap"
)

const Mock = "mock_receiver"

type MockReceiver struct {
	cfg             *Config
	analyzerManager analyzerpackage.Manager
	logger          *zap.Logger
}

func NewMockReceiver(cfg interface{}, logger *zap.Logger, analyzerManager analyzerpackage.Manager) Receiver {
	config, ok := cfg.(*Config)
	if !ok {
		logger.Sugar().Panicf("Cannot convert mock_analyzer config")
	}
	return &MockReceiver{
		cfg:             config,
		analyzerManager: analyzerManager,
		logger:          logger,
	}
}

func (r *MockReceiver) Start() error {
	r.logger.Sugar().Infof("Start MockReceiver...")
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
			r.logger.Sugar().Infof("Failed to consume event: %s, error is: %v", events, err)
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
