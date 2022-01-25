package otelexporter

import (
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.uber.org/zap"
	"net/http"
)

func StartServer(exporter *prometheus.Exporter, logger *zap.Logger, port string) error {
	http.HandleFunc("/metrics", exporter.ServeHTTP)

	srv := http.Server{
		Addr:    port,
		Handler: http.DefaultServeMux,
	}

	logger.Sugar().Infof("Prometheus Server listening at port: [%s]", port)
	err := srv.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	logger.Sugar().Infof("Prometheus gracefully shutdown the http server...\n")

	return nil
}
