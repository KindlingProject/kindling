package esexporter

import (
	"context"
	"log"
	"os"

	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/olivere/elastic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Type = "esexporter"

type EsExporter struct {
	esClient *elastic.Client

	telemetry *component.TelemetryTools
	config    *Config
}

func New(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	cfg, _ := config.(*Config)
	ret := &EsExporter{
		config:    cfg,
		telemetry: telemetry,
	}
	errorLog := log.New(os.Stdout, "app", log.LstdFlags)
	var err error
	ret.esClient, err = elastic.NewClient(elastic.SetErrorLog(errorLog), elastic.SetURL(cfg.GetEsHost()), elastic.SetSniff(false))
	if err != nil {
		telemetry.Logger.Error("new es client error", zap.Error(err))
	}
	return ret
}

func (e *EsExporter) Consume(dataGroup *model.DataGroup) error {
	if ce := e.telemetry.Logger.Check(zapcore.DebugLevel, "Receiver DataGroup"); ce != nil {
		ce.Write(
			zap.String("dataGroup", dataGroup.String()),
		)
	}
	if dataGroup == nil {
		// no need consume
		return nil
	}

	switch dataGroup.Name {
	case constnames.AggregatedNetRequestMetricGroup:
		// We don't care about metrics now.
		return nil
	case constnames.SingleNetRequestMetricGroup:
		e.sendTrace(dataGroup)
		return nil
	}
	return nil
}

func (e *EsExporter) sendTrace(dataGroup *model.DataGroup) {
	e.telemetry.Logger.Info("Will send a trace to ElasticSearch")
	if e.esClient == nil {
		return
	}
	trace := TraceData{
		Name:      dataGroup.Name,
		Metrics:   make(map[string]int64),
		Labels:    dataGroup.Labels.ToStringMap(),
		Timestamp: dataGroup.Timestamp,
	}
	for _, metric := range dataGroup.Metrics {
		trace.Metrics[metric.Name] = metric.GetInt().Value
	}
	e.esClient.Index().Index("kindling_trace").Type("_doc").BodyJson(trace).Do(context.Background())
}

type TraceData struct {
	Name      string            `json:"name"`
	Metrics   map[string]int64  `json:"metrics"`
	Labels    map[string]string `json:"labels"`
	Timestamp uint64            `json:"timestamp"`
}
