package controller

import (
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

func TestModule(t *testing.T) {
	sample := NewModule("profile", component.NewTelemetryManager().GetGlobalTelemetryTools(), Stopped)

	stopSignal := make(chan struct{})
	go sample.Start(
		WithStopInterval(5*time.Second),
		WithStopSignal(stopSignal),
	)

	timer := time.NewTimer(6 * time.Second)
	<-timer.C
	close(stopSignal)

	<-time.After(10 * time.Second)
}
