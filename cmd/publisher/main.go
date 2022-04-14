package main

import (
	"context"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	ingester "github.com/frain-dev/convoy-ingester"
)

func main() {
	ctx := context.Background()
	if err := funcframework.RegisterEventFunctionContext(ctx, "/", ingester.PushToConvoy); err != nil {
		log.Printf("EventFunction: %v\n", err)
	}

	// Use PORT environment variable, or default to 8080.
	port := "8090"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
