package main

import (
	"log"

	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/models"
	"github.com/cohesion-dev/GNX/backend/pkg/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbCfg := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
	}

	db, err := database.InitDatabase(dbCfg)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&models.Comic{},
		&models.ComicRole{},
		&models.ComicSection{},
		&models.ComicStoryboardPage{},
		&models.ComicStoryboardPanel{},
		&models.ComicStoryboardSegment{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully!")
}
