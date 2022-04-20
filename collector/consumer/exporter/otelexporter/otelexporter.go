package otelexporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kindling-project/kindling/collector/component"
	"github.com/Kindling-project/kindling/collector/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/model"
	"github.com/Kindling-project/kindling/collector/model/constlabels"
	"github.com/Kindling-project/kindling/collector/model/constnames"
	"github.com/Kindling-project/kindling/collector/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	exportmetric "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	apitrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"os"
	"time"
)

const (
	Otel                    = "otelexporter"
	StdoutKindExporter      = "stdout"
	OtlpGrpcKindExporter    = "otlp"
	PrometheusKindExporter  = "prometheus"
	Int64BoundaryMultiplier = 1e6

	MeterName  = "kindling-instrument"
	TracerName = "kindling-tracer"
)

var serviceName string

type labelKey struct {
	metric          string
	srcIp           string
	dstIp           string
	dstPort         int64
	requestContent  string
	responseContent string
	statusCode      string
	protocol        string
}

type OtelOutputExporters struct {
	metricExporter exportmetric.Exporter
	traceExporter  sdktrace.SpanExporter
}

type OtelExporter struct {
	cfg                  *Config
	metricController     *controller.Controller
	traceProvider        *sdktrace.TracerProvider
	defaultTracer        trace.Tracer
	metricAggregationMap map[string]MetricAggregationKind
	customLabels         []attribute.KeyValue
	instrumentFactory    *instrumentFactory
	telemetry            *component.TelemetryTools
	adapterManager       *BaseAdapterManager
}

func NewExporter(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	newSelfMetrics(telemetry.MeterProvider)
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert Component config", zap.String("componentType", Otel))
	}
	customLabels := make([]attribute.KeyValue, 0, len(cfg.CustomLabels))
	for k, v := range cfg.CustomLabels {
		customLabels = append(customLabels, attribute.String(k, v))
	}

	if cfg.ExportKind != PrometheusKindExporter {
		commonLabels := GetCommonLabels(false, telemetry.Logger)
		for i := 0; i < len(commonLabels); i++ {
			if _, find := cfg.CustomLabels[string(commonLabels[i].Key)]; !find {
				customLabels = append(customLabels, commonLabels[i])
			}
		}
	}

	hostName, err := os.Hostname()
	if err != nil {
		telemetry.Logger.Error("Error happened when getting hostname; set hostname unknown: ", zap.Error(err))
		hostName = "unknown"
	}

	clusterId, ok := os.LookupEnv(clusterIdEnv)
	if !ok {
		telemetry.Logger.Warn("[CLUSTER_ID] is not found in env variable which will be set [noclusteridset]")
		clusterId = "noclusteridset"
	}

	serviceName = CmonitorServiceNamePrefix + "-" + clusterId
	rs, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceInstanceIDKey.String(hostName),
		),
	)

	var otelexporter *OtelExporter
	var cont *controller.Controller

	if cfg.ExportKind == PrometheusKindExporter {
		config := prometheus.Config{}
		// Create a meter
		c := controller.New(
			otelprocessor.NewFactory(
				selector.NewWithHistogramDistribution(
					histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
				),
				aggregation.CumulativeTemporalitySelector(),
			),
			controller.WithResource(rs),
		)
		exp, err := prometheus.New(config, c)

		if err != nil {
			telemetry.Logger.Panic("failed to initialize prometheus exporter %v", zap.Error(err))
			return nil
		}

		otelexporter = &OtelExporter{
			cfg:                  cfg,
			metricController:     c,
			traceProvider:        nil,
			defaultTracer:        nil,
			customLabels:         customLabels,
			instrumentFactory:    newInstrumentFactory(exp.MeterProvider().Meter(MeterName), telemetry.Logger, customLabels),
			metricAggregationMap: cfg.MetricAggregationMap,
			telemetry:            telemetry,
			adapterManager:       createBaseAdapterManager(customLabels),
		}
		go func() {
			err := StartServer(exp, telemetry.Logger, cfg.PromCfg.Port)
			if err != nil {
				telemetry.Logger.Warn("error starting otelexporter prometheus server: ", zap.Error(err))
			}
		}()
	} else {
		var collectPeriod time.Duration

		if cfg.ExportKind == StdoutKindExporter {
			collectPeriod = cfg.StdoutCfg.CollectPeriod
		} else if cfg.ExportKind == OtlpGrpcKindExporter {
			collectPeriod = cfg.OtlpGrpcCfg.CollectPeriod
		} else {
			telemetry.Logger.Panic("Err! No exporter kind matched ", zap.String("exportKind", cfg.ExportKind))
			return nil
		}

		exporters, err := newExporters(context.Background(), cfg, telemetry.Logger)
		if err != nil {
			telemetry.Logger.Panic("Error happened when creating otel exporter:", zap.Error(err))
			return nil
		}

		cont = controller.New(
			otelprocessor.NewFactory(simple.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
			), exporters.metricExporter),
			controller.WithExporter(exporters.metricExporter),
			controller.WithCollectPeriod(collectPeriod),
			controller.WithResource(rs),
		)

		// Init TraceProvider
		ssp := sdktrace.NewBatchSpanProcessor(
			exporters.traceExporter,
			sdktrace.WithMaxQueueSize(2048),
			sdktrace.WithMaxExportBatchSize(512),
		)

		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(ssp),
			sdktrace.WithResource(rs),
		)

		tracer := tracerProvider.Tracer(TracerName)

		otelexporter = &OtelExporter{
			cfg:                  cfg,
			metricController:     cont,
			traceProvider:        tracerProvider,
			defaultTracer:        tracer,
			customLabels:         customLabels,
			instrumentFactory:    newInstrumentFactory(cont.Meter(MeterName), telemetry.Logger, customLabels),
			metricAggregationMap: cfg.MetricAggregationMap,
			telemetry:            telemetry,
			adapterManager:       createBaseAdapterManager(customLabels),
		}

		if err = cont.Start(context.Background()); err != nil {
			telemetry.Logger.Panic("failed to start controller:", zap.Error(err))
			return nil
		}
	}

	return otelexporter
}

func (e *OtelExporter) Consume(gaugeGroup *model.GaugeGroup) error {
	if gaugeGroup == nil {
		// no need consume
		return nil
	}
	gaugeGroupReceiverCounter.Add(context.Background(), 1, attribute.String("name", gaugeGroup.Name))
	if ce := e.telemetry.Logger.Check(zap.DebugLevel, "exporter receives a gaugeGroup: "); ce != nil {
		ce.Write(
			zap.String("gaugeGroup", gaugeGroup.String()),
		)
	}
	if gaugeGroup.Name == constnames.AggregatedNetRequestGaugeGroup {
		e.PushNetMetric(gaugeGroup)
		return nil
	} else if gaugeGroup.Name == constnames.SingleNetRequestGaugeGroup {
		if e.defaultTracer != nil && e.cfg.AdapterConfig.NeedTraceAsResourceSpan {
			attrs, _ := e.adapterManager.traceToSpanAdapter.adapt(gaugeGroup)
			_, span := e.defaultTracer.Start(
				context.Background(),
				constvalues.SpanInfo,
				apitrace.WithAttributes(attrs...),
			)
			span.End()
		} else if e.defaultTracer != nil && e.cfg.AdapterConfig.NeedTraceAsResourceSpan {
			return errors.New("send span failed: this exporter can not support Span Data")
		}

		if e.cfg.AdapterConfig.NeedTraceAsMetric {
			attrsMap, _ := e.adapterManager.traceToMetricAdapter.transform(gaugeGroup)
			var requestTotalTime *model.Gauge
			for i := 0; i < len(gaugeGroup.Values); i++ {
				if gaugeGroup.Values[i].Name == constvalues.RequestTotalTime {
					requestTotalTime = &model.Gauge{
						Name:  constlabels.ToKindlingTraceAsMetricName(),
						Value: gaugeGroup.Values[i].Value,
					}
					break
				}
			}
			e.instrumentFactory.recordLastValue(constlabels.ToKindlingTraceAsMetricName(), &model.GaugeGroup{
				Values:    []*model.Gauge{requestTotalTime},
				Labels:    attrsMap,
				Timestamp: gaugeGroup.Timestamp,
			})
		}
		return nil
	}

	// For Other metric
	if gaugeGroup.Name == constvalues.SpanInfo {
		if e.defaultTracer == nil {
			return errors.New("send span failed: this exporter can not support Span Data")
		}
		return e.PushTrace(gaugeGroup, gaugeGroup.Name)
	} else {
		return e.PushMetric(gaugeGroup)
	}
}

func (e *OtelExporter) PushNetMetric(gaugeGroup *model.GaugeGroup) {
	isServer := gaugeGroup.Labels.GetBoolValue(constlabels.IsServer)
	if e.cfg.AdapterConfig.StoreExternalSrcIP {
		srcNamespace := gaugeGroup.Labels.GetStringValue(constlabels.SrcNamespace)
		if srcNamespace == constlabels.ExternalClusterNamespace && isServer {
			e.RecordNetMetric(e.adapterManager.detailTopologyAdapter, gaugeGroup, false)
		}
	}
	var metricAdapter [2]*Adapter
	if e.cfg.AdapterConfig.NeedPodDetail {
		if isServer {
			metricAdapter = e.adapterManager.detailEntityAdapter
		} else {
			metricAdapter = e.adapterManager.detailTopologyAdapter
		}
	} else {
		if isServer {
			metricAdapter = e.adapterManager.aggEntityAdapter
		} else {
			metricAdapter = e.adapterManager.aggTopologyAdapter
		}
	}

	e.RecordNetMetric(metricAdapter, gaugeGroup, isServer)
}

func (e *OtelExporter) PushMetric(gaugeGroup *model.GaugeGroup) error {
	//storeGaugeGroupKeys(gaugeGroup)
	values := gaugeGroup.Values
	measurements := make([]metric.Measurement, 0, len(values))
	for _, value := range values {
		num := value.Value
		name := value.Name
		metricKind, ok := e.findInstrumentKind(name)
		if !ok {
			e.telemetry.Logger.Warn("Skip a Metric: No metric aggregation set for metric", zap.String("metricName", name))
			continue
		}
		if metricKind == MAGaugeKind {
			e.instrumentFactory.recordLastValue(name, &model.GaugeGroup{
				Values:    []*model.Gauge{value},
				Labels:    gaugeGroup.Labels,
				Timestamp: gaugeGroup.Timestamp})
		} else {
			measurements = append(measurements, e.instrumentFactory.getInstrument(name, metricKind).Measurement(num))
		}
	}
	if len(measurements) > 0 {
		labels := gaugeGroup.Labels
		e.instrumentFactory.meter.RecordBatch(context.Background(), GetLabels(labels, e.customLabels), measurements...)
	}
	return nil
}

func (e *OtelExporter) PushTrace(g *model.GaugeGroup, spanName string) error {
	_, span := e.defaultTracer.Start(
		context.Background(),
		spanName,
		apitrace.WithAttributes(GetLabels(g.Labels, e.customLabels)...),
	)
	span.End()
	return nil
}

func (e *OtelExporter) findInstrumentKind(metricName string) (MetricAggregationKind, bool) {
	kind, find := e.metricAggregationMap[metricName]
	return kind, find
}

func ToStringKeyValues(values map[string]model.AttributeValue) []attribute.KeyValue {
	stringKeyValues := make([]attribute.KeyValue, 0, len(values))
	for k, v := range values {
		stringKeyValues = append(stringKeyValues, attribute.String(k, v.ToString()))
	}
	return stringKeyValues
}

// Crete new opentelemetry-go exporter.
func newExporters(context context.Context, cfg *Config, logger *zap.Logger) (*OtelOutputExporters, error) {
	var retExporters *OtelOutputExporters
	logger.Sugar().Infof("Initializing OpenTelemetry exporter whose type is %s", cfg.ExportKind)
	switch cfg.ExportKind {
	case StdoutKindExporter:
		metricExp, err := stdoutmetric.New(
			stdoutmetric.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		traceExp, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		retExporters = &OtelOutputExporters{
			metricExporter: metricExp,
			traceExporter:  traceExp,
		}
	case OtlpGrpcKindExporter:
		metricExporter, err := otlpmetricgrpc.New(context,
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpoint(cfg.OtlpGrpcCfg.Endpoint),
			otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetrySettings{
				Enabled:         true,
				InitialInterval: 300 * time.Millisecond,
				MaxInterval:     5 * time.Second,
				MaxElapsedTime:  15 * time.Second,
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		traceExporter, err := otlptracegrpc.New(context,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(cfg.OtlpGrpcCfg.Endpoint),
			otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
				Enabled:         true,
				InitialInterval: 300 * time.Millisecond,
				MaxInterval:     5 * time.Second,
				MaxElapsedTime:  15 * time.Second,
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create exporter, %w", err)
		}
		retExporters = &OtelOutputExporters{
			metricExporter: metricExporter,
			traceExporter:  traceExporter,
		}
	default:
		return nil, errors.New("failed to create exporter, no exporter kind is provided")
	}
	return retExporters, nil
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

func (e *OtelExporter) RecordNetMetric(adapter [2]*Adapter, group *model.GaugeGroup, isServer bool) {
	requestCount, err := e.GetMetricMeasurementOnlyRequestCount(group, isServer)
	if err == nil {
		attrsWithIsSlow, _ := adapter[1].adapt(group)
		e.instrumentFactory.meter.RecordBatch(context.Background(), attrsWithIsSlow, requestCount)
	} else {
		e.telemetry.Logger.Error("Can not record Metric", zap.Error(err))
	}
	attrs, _ := adapter[0].adapt(group)
	e.instrumentFactory.meter.RecordBatch(context.Background(), attrs, e.GetMetricMeasurementExceptRequestCount(group, false)...)
}

func (e *OtelExporter) GetMetricMeasurementExceptRequestCount(gaugeGroup *model.GaugeGroup, isServer bool) []metric.Measurement {
	gauges := gaugeGroup.Values
	measurements := make([]metric.Measurement, 0, len(gauges))
	for i := 0; i < len(gauges); i++ {
		name := constlabels.ToKindlingMetricName(gauges[i].Name, isServer)
		if name == "" || gauges[i].Name == constvalues.RequestCount {
			continue
		}
		if gauges[i].Value < 0 {
			e.telemetry.Logger.Warn("Exporter received an negative value!", zap.String("gauge", gaugeGroup.String()))
			continue
		}
		measurements = append(measurements, e.instrumentFactory.getInstrument(name, e.metricAggregationMap[name]).Measurement(gauges[i].Value))
	}
	return measurements
}

func (e *OtelExporter) GetMetricMeasurementOnlyRequestCount(gaugeGroup *model.GaugeGroup, isServer bool) (metric.Measurement, error) {
	gauges := gaugeGroup.Values
	for i := 0; i < len(gauges); i++ {
		if gauges[i].Name == constvalues.RequestCount && gauges[i].Value >= 0 {
			name := constlabels.ToKindlingMetricName(gauges[i].Name, isServer)
			return e.instrumentFactory.getInstrument(name, e.metricAggregationMap[name]).Measurement(gauges[i].Value), nil
		}
	}
	return metric.Measurement{}, errors.New("no requestCount or requestCount is a negative value")
}
