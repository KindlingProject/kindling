package observability

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.uber.org/zap"
	"net/http"
	"os"
)

const (
	clusterIdEnv              = "CLUSTER_ID"
	userIdEnv                 = "USER_ID"
	regionIdEnv               = "REGION_ID"
	KindlingServiceNamePrefix = "kindling"
)

func InitTelemetry(logger *zap.Logger, config *Config) (metric.MeterProvider, error) {
	hostName, err := os.Hostname()
	if err != nil {
		logger.Error("Error happened when getting hostname; set hostname unknown: ", zap.Error(err))
		hostName = "unknown"
	}

	clusterId, ok := os.LookupEnv(clusterIdEnv)
	if !ok {
		logger.Warn("[CLUSTER_ID] is not found in env variable which will be set [noclusteridset]")
		clusterId = "noclusteridset"
	}
	serviceName := KindlingServiceNamePrefix + "-" + clusterId
	rs, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceInstanceIDKey.String(hostName),
		),
	)

	promConfig := prometheus.Config{
		DefaultHistogramBoundaries: exponentialInt64NanosecondsBoundaries,
	}
	// Create controller
	c := controller.New(
		otelprocessor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(promConfig.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
		),
		controller.WithResource(rs),
	)
	exp, err := prometheus.New(promConfig, c)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize self-telemetry prometheus %w", err)
	}

	go StartServer(exp, logger, config.Port)
	if err = c.Start(context.Background()); err != nil {
		return nil, err
	}
	return exp.MeterProvider(), nil
}

func StartServer(exporter *prometheus.Exporter, logger *zap.Logger, port string) error {
	serveMux := http.ServeMux{}
	serveMux.HandleFunc("/metrics", exporter.ServeHTTP)

	srv := http.Server{
		Addr:    port,
		Handler: &serveMux,
	}

	logger.Sugar().Infof("Prometheus Server listening at port: [%s]", port)
	err := srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	logger.Sugar().Infof("Prometheus gracefully shutdown the http server...\n")

	return nil
}
