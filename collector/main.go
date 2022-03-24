package main

import (
	"github.com/Kindling-project/kindling/collector/application"
	"log"
	"fmt"

	"github.com/sugary199/collector-version/core"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	
	//print version information
	fmt.Printf("%s\n", core.Version())
	log.Printf("GitCommitInfo:%s\n", core.Version())
	log.Printf("-------------")
	log.Printf("-------------")
	log.Printf("-------------")
	log.Printf("-------------")

	
	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}

}
