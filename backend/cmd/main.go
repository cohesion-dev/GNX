package main

import (
	"log"

	"github.com/cohesion-dev/GNX/backend_new/internal/app"
)

func main() {
	application := app.NewApp()

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
