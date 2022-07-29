package component

import (
	"log"

	"github.com/Kindling-project/kindling/collector/pkg/observability"
	"github.com/Kindling-project/kindling/collector/pkg/observability/logger"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ObservabilityConfig = "observability"
	LogKey              = "logger"
	MetricKey           = "opentelemetry"
)

type TelemetryManager struct {
	MeterProvider metric.MeterProvider
	Logger        *zap.Logger
	Selector      []string
}

func NewTelemetryManager() *TelemetryManager {
	return &TelemetryManager{
		MeterProvider: metric.NewNoopMeterProvider(),
		Logger:        logger.CreateDefaultLogger(),
		Selector:      make([]string, 0),
	}
}

func (t *TelemetryManager) ConstructConfig(viper *viper.Viper) {
	t.initLogger(viper)
	t.initProvider(viper)
}

func (t *TelemetryManager) initLogger(viper *viper.Viper) {
	var loggerConfig = logger.Config{}
	key := ObservabilityConfig + "." + LogKey
	err := viper.UnmarshalKey(key, &loggerConfig)
	if err != nil {
		log.Printf("Error happened when reading logger config, and default config will be used: %v", err)
	}
	t.Selector = loggerConfig.Selector
	t.Logger = logger.InitLogger(loggerConfig)
}

func (t *TelemetryManager) initProvider(viper *viper.Viper) {
	var config = &observability.DefaultConfig
	key := ObservabilityConfig + "." + MetricKey
	err := viper.UnmarshalKey(key, config)
	if err != nil {
		log.Printf("Error happened when reading observability config, and default config will be used: %v", err)
	}
	meterProvider, err := observability.InitTelemetry(t.Logger, config)
	if err != nil {
		log.Printf("Error happened when initializing meter provider: %v", err)
		return
	}
	t.MeterProvider = meterProvider
	// Here we set a global meter provider to make it accessible to all components
	global.SetMeterProvider(meterProvider)
}

func (t *TelemetryManager) GetGlobalTelemetryTools() *TelemetryTools {
	return t.getToolsWithOption(
		WithDebug(true),
	)
}

func (t *TelemetryManager) GetTelemetryTools(component string) *TelemetryTools {
	options := make([]TelemetryOption, 0)
	// DebugSelector
	options = append(options, t.SelectorOption(component))

	return t.getToolsWithOption(options...)
}

func (t *TelemetryManager) SelectorOption(component string) TelemetryOption {
	if len(t.Selector) == 0 {
		return WithDebug(true)
	}
	for _, selectedComponent := range t.Selector {
		if selectedComponent == component {
			return WithDebug(true)
		}
	}
	return WithDebug(false)
}

func (t *TelemetryManager) getToolsWithOption(opts ...TelemetryOption) *TelemetryTools {
	newSubTool := &TelemetryTools{
		MeterProvider: t.MeterProvider,
		Logger: &TelemetryLogger{
			Logger: t.Logger,
		},
	}

	for _, opt := range opts {
		opt(newSubTool)
	}

	return newSubTool
}

type TelemetryLogger struct {
	*zap.Logger
	EnableDebug bool
}

type TelemetryTools struct {
	MeterProvider metric.MeterProvider
	Logger        *TelemetryLogger
}

func (t *TelemetryTools) GetZapLogger() *zap.Logger {
	return t.Logger.Logger
}

func (t *TelemetryLogger) Debug(msg string, fields ...zap.Field) {
	if t.EnableDebug {
		t.Logger.Debug(msg, fields...)
	}
}

func (t *TelemetryLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	if t.EnableDebug {
		return t.Logger.Check(lvl, msg)
	} else {
		return nil
	}
}

type TelemetryOption func(*TelemetryTools)

func WithDebug(enableDebug bool) TelemetryOption {
	return func(tt *TelemetryTools) {
		tt.Logger.EnableDebug = enableDebug
	}
}

func NewDefaultTelemetryTools() *TelemetryTools {
	return &TelemetryTools{
		Logger: &TelemetryLogger{
			Logger:      logger.CreateDefaultLogger(),
			EnableDebug: true,
		},
		MeterProvider: metric.NewNoopMeterProvider(),
	}
}
