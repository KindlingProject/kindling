package prometheusexporter

import (
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/consumer/exporter/otelexporter/defaultadapter"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net"
	"net/http"
)

const TYPE = "prometheus"

const (
	Int64BoundaryMultiplier = 1e6
)

var serviceName string

type prometheusExporter struct {
	cfg                  *Config
	metricAggregationMap map[string]MetricAggregationKind
	customLabels         []attribute.KeyValue
	telemetry            *component.TelemetryTools
	collector            *collector
	registry             *prometheus.Registry
	shutdownFunc         func() error
	handler              http.Handler

	adapters []defaultadapter.Adapter
}

func NewExporter(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert Component config", zap.String("componentType", TYPE))
	}

	collector := newCollector(cfg, telemetry.Logger)
	registry := prometheus.NewRegistry()
	registry.Register(collector)

	prometheusExporter := &prometheusExporter{
		cfg:                  cfg,
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            telemetry,
		adapters: []defaultadapter.Adapter{
			defaultadapter.NewNetAdapter(nil, &defaultadapter.NetAdapterConfig{
				StoreTraceAsMetric: cfg.AdapterConfig.NeedTraceAsMetric,
				StoreTraceAsSpan:   cfg.AdapterConfig.NeedTraceAsResourceSpan,
				StorePodDetail:     cfg.AdapterConfig.NeedPodDetail,
				StoreExternalSrcIP: cfg.AdapterConfig.StoreExternalSrcIP,
			}),
			defaultadapter.NewSimpleAdapter([]string{constnames.TcpGaugeGroupName}, nil),
		},
	}

	go func() {
		prometheusExporter.Start(context.Background())
	}()
	return prometheusExporter
}

func (p *prometheusExporter) findInstrumentKind(metricName string) (MetricAggregationKind, bool) {
	kind, find := p.metricAggregationMap[metricName]
	return kind, find
}

var exponentialInt64Boundaries = []float64{10, 25, 50, 80, 130, 200, 300,
	400, 500, 700, 1000, 2000, 5000, 30000}

// exponentialInt64NanoSecondsBoundaries applies a multiplier to the exponential
// Int64Boundaries: [ 5M, 10M, 20M, 40M, ...]
var exponentialInt64NanosecondsBoundaries = func(bounds []float64) (asint []float64) {
	for _, f := range bounds {
		asint = append(asint, Int64BoundaryMultiplier*f)
	}
	return
}(exponentialInt64Boundaries)

func (p *prometheusExporter) Start(_ context.Context) error {
	ln, err := net.Listen("tcp", p.cfg.PromCfg.Port)
	if err != nil {
		return err
	}

	p.shutdownFunc = ln.Close

	mux := http.NewServeMux()
	mux.Handle("/metrics", p.handler)
	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(ln)
	}()

	return nil
}
