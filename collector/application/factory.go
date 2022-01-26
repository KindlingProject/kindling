package application

import (
	"github.com/Kindling-project/kindling/collector/analyzer"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/consumer/processor"
	"github.com/Kindling-project/kindling/collector/receiver"
	"github.com/spf13/viper"
)

const (
	ReceiversKey  = "receivers"
	AnalyzersKey  = "analyzers"
	ProcessorsKey = "processors"
	ExportersKey  = "exporters"
)

var ComponentsKeyMap = []string{ReceiversKey, AnalyzersKey, ProcessorsKey, ExportersKey}

type ComponentsFactory struct {
	Receivers  map[string]ReceiverFactory
	Analyzers  map[string]AnalyzerFactory
	Processors map[string]ProcessorFactory
	Exporters  map[string]ExporterFactory
}

type NewReceiverFunc func(cfg interface{}, telemetry *component.TelemetryTools, analyzerManager analyzer.Manager) receiver.Receiver
type NewAnalyzerFunc func(cfg interface{}, telemetry *component.TelemetryTools, consumers []consumer.Consumer) analyzer.Analyzer
type NewProcessorFunc func(cfg interface{}, telemetry *component.TelemetryTools, consumer consumer.Consumer) processor.Processor
type NewExporterFunc func(cfg interface{}, telemetry *component.TelemetryTools) exporter.Exporter

type ReceiverFactory struct {
	NewFunc NewReceiverFunc
	Config  interface{}
}

type AnalyzerFactory struct {
	NewFunc NewAnalyzerFunc
	Config  interface{}
}

type ProcessorFactory struct {
	NewFunc NewProcessorFunc
	Config  interface{}
}

type ExporterFactory struct {
	NewFunc NewExporterFunc
	Config  interface{}
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
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterAnalyzer(
	name string,
	f NewAnalyzerFunc,
	config interface{},
) {
	c.Analyzers[name] = AnalyzerFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterProcessor(
	name string,
	f NewProcessorFunc,
	config interface{},
) {
	c.Processors[name] = ProcessorFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) RegisterExporter(
	name string,
	f NewExporterFunc,
	config interface{},
) {
	c.Exporters[name] = ExporterFactory{
		NewFunc: f,
		Config:  config,
	}
}

func (c *ComponentsFactory) ConstructConfig(viper *viper.Viper) error {
	for _, componentKind := range ComponentsKeyMap {
		switch componentKind {
		case ReceiversKey:
			for k, factory := range c.Receivers {
				key := ReceiversKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config)
				if err != nil {
					return err
				}
			}
		case AnalyzersKey:
			for k, factory := range c.Analyzers {
				key := AnalyzersKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config)
				if err != nil {
					return err
				}
			}
		case ProcessorsKey:
			for k, factory := range c.Processors {
				key := ProcessorsKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config)
				if err != nil {
					return err
				}
			}
		case ExportersKey:
			for k, factory := range c.Exporters {
				key := ExportersKey + "." + k
				err := viper.UnmarshalKey(key, factory.Config)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
