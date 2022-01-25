package application

import (
	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/logger"
	"github.com/Kindling-project/kindling/collector/receiver"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	ReceiversKey  = "receivers"
	AnalyzersKey  = "analyzers"
	ProcessorsKey = "processors"
	ExportersKey  = "exporters"
	LoggerKey     = "logger"
)

var ComponentsKeyMap = []string{ReceiversKey, AnalyzersKey, ProcessorsKey, ExportersKey, LoggerKey}

type ComponentsFactory struct {
	Receivers  map[string]ReceiverFactory
	Analyzers  map[string]AnalyzerFactory
	Processors map[string]ProcessorFactory
	Exporters  map[string]ExporterFactory
	logger     *zap.Logger
}

type NewReceiverFunc func(cfg interface{}, log *zap.Logger, analyzerManager analyzer.Manager) receiver.Receiver
type NewAnalyzerFunc func(cfg interface{}, log *zap.Logger, consumers []consumer.Consumer) analyzer.Analyzer
type NewProcessorFunc func(cfg interface{}, log *zap.Logger, consumer consumer.Consumer) processor.Processor
type NewExporterFunc func(cfg interface{}, log *zap.Logger) exporter.Exporter

type ReceiverFactory struct {
	newFunc NewReceiverFunc
	config  interface{}
}

type AnalyzerFactory struct {
	newFunc NewAnalyzerFunc
	config  interface{}
}

type ProcessorFactory struct {
	newFunc NewProcessorFunc
	config  interface{}
}

type ExporterFactory struct {
	newFunc NewExporterFunc
	config  interface{}
}

func NewComponentsFactory() *ComponentsFactory {
	return &ComponentsFactory{
		Receivers:  make(map[string]ReceiverFactory),
		Analyzers:  make(map[string]AnalyzerFactory),
		Processors: make(map[string]ProcessorFactory),
		Exporters:  make(map[string]ExporterFactory),
	}
}
func (c *ComponentsFactory) RegisterReceiver(
	name string,
	f NewReceiverFunc,
	config interface{},
) {
	c.Receivers[name] = ReceiverFactory{
		newFunc: f,
		config:  config,
	}
}

func (c *ComponentsFactory) RegisterAnalyzer(
	name string,
	f NewAnalyzerFunc,
	config interface{},
) {
	c.Analyzers[name] = AnalyzerFactory{
		newFunc: f,
		config:  config,
	}
}

func (c *ComponentsFactory) RegisterProcessor(
	name string,
	f NewProcessorFunc,
	config interface{},
) {
	c.Processors[name] = ProcessorFactory{
		newFunc: f,
		config:  config,
	}
}

func (c *ComponentsFactory) RegisterExporter(
	name string,
	f NewExporterFunc,
	config interface{},
) {
	c.Exporters[name] = ExporterFactory{
		newFunc: f,
		config:  config,
	}
}

func (c *ComponentsFactory) ConstructConfig(viper *viper.Viper) error {
	for _, componentKind := range ComponentsKeyMap {
		switch componentKind {
		case ReceiversKey:
			for k, factory := range c.Receivers {
				key := ReceiversKey + "." + k
				err := viper.UnmarshalKey(key, factory.config)
				if err != nil {
					return err
				}
			}
		case AnalyzersKey:
			for k, factory := range c.Analyzers {
				key := AnalyzersKey + "." + k
				err := viper.UnmarshalKey(key, factory.config)
				if err != nil {
					return err
				}
			}
		case ProcessorsKey:
			for k, factory := range c.Processors {
				key := ProcessorsKey + "." + k
				err := viper.UnmarshalKey(key, factory.config)
				if err != nil {
					return err
				}
			}
		case ExportersKey:
			for k, factory := range c.Exporters {
				key := ExportersKey + "." + k
				err := viper.UnmarshalKey(key, factory.config)
				if err != nil {
					return err
				}
			}
		case LoggerKey:
			var loggerConfig = logger.Config{}
			err := viper.UnmarshalKey(LoggerKey, &loggerConfig)
			if err != nil {
				return err
			}
			c.logger = logger.InitLogger(loggerConfig)
		}
	}
	return nil
}
