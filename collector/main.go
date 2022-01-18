package main

import (
	"github.com/Kindling-project/kindling/collector/application"
	"log"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
