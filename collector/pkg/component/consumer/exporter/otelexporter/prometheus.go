package otelexporter

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

var (
	mu   sync.Mutex
	srv  *http.Server
)

func StartServer(exporter *prometheus.Exporter, telemetry *component.TelemetryTools, port string) error {
	mu.Lock()

	if srv != nil {
		if err := srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", exporter.ServeHTTP)

	srv = &http.Server{
		Addr:    port,
		Handler: mux,
	}

	telemetry.Logger.Infof("Prometheus Server listening at port: [%s]", port)

	go func() {
		mu.Unlock()
		srv.ListenAndServe()
		telemetry.Logger.Infof("Prometheus gracefully shutdown the http server...\n")
	}()

	return nil
}
