// Command server runs the gRPC API. It wires every domain through
// app.App (Initialize -> Bootstrap -> Start -> AwaitShutdown) and
// shuts down gracefully on SIGINT or SIGTERM.
package main

import (
	"log"

	"github.com/isyll/go-api-starter/internal/app"
)

func main() {
	application := app.New()

	if err := application.Initialize(); err != nil {
		log.Fatalf("initialization failed: %v", err)
	}
	if err := application.Bootstrap(); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	application.Start()
	application.AwaitShutdown()
}
