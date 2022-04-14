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
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", ingester.WebhookEndpoint); err != nil {
		log.Printf("HTTPFunction: %v\n", err)
	}

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
