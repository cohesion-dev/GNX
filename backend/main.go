package main

import (
	"log"

	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/app"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := config.SetupDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	server := app.NewServer(db, cfg)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
