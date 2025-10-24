package main

import (
	"log"

	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/models"
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

	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&models.Comic{},
		&models.ComicRole{},
		&models.ComicSection{},
		&models.ComicStoryboard{},
		&models.ComicStoryboardDetail{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully!")
}
