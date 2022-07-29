package otelexporter

import (
	"net/http"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

func StartServer(exporter *prometheus.Exporter, telemetry *component.TelemetryTools, port string) error {
	http.HandleFunc("/metrics", exporter.ServeHTTP)

	srv := http.Server{
		Addr:    port,
		Handler: http.DefaultServeMux,
	}

	telemetry.Logger.Sugar().Infof("Prometheus Server listening at port: [%s]", port)
	err := srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	telemetry.Logger.Sugar().Infof("Prometheus gracefully shutdown the http server...\n")

	return nil
}
