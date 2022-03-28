package main

import (
	"github.com/Kindling-project/kindling/collector/application"
	"log"
//  package in formal kindlingproject if merged:
//  "github.com/Kindling-project/kindling/collector/version"
	"github.com/sugary199/collector-version/version"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	
	//print version information
	log.Printf("GitCommitInfo:%s\n", version.Version())

	
	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}

}
