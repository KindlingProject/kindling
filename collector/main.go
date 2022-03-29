package main

import (
	"github.com/Kindling-project/kindling/collector/application"
	"github.com/Kindling-project/kindling/collector/version"
	"log"
)

func main() {
	//print version information
	log.Printf("GitCommitInfo:%s\n", version.Version())

	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}

}
