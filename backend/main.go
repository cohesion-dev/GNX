package main

import (
	"log"

	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/app"
	"github.com/cohesion-dev/GNX/backend/pkg/database"
	"github.com/gin-gonic/gin"
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
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := gin.Default()
	
	server := app.NewServer(db)
	server.SetupRoutes(router)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
