package app

import (
	"fmt"
	"log"

	"github.com/cohesion-dev/GNX/backend_new/config"
	"github.com/cohesion-dev/GNX/backend_new/internal/handlers"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/internal/services"
	"github.com/cohesion-dev/GNX/backend_new/pkg/aigc"
	"github.com/cohesion-dev/GNX/backend_new/pkg/database"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
)

type App struct {
	config *config.Config
	router *Router
}

func NewApp() *App {
	cfg := config.Load()

	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	storageClient := storage.NewStorage(&cfg.Storage)

	aigcClient := aigc.NewAIGC(&cfg.AI)

	comicRepo := repositories.NewComicRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	sectionRepo := repositories.NewSectionRepository(db)
	pageRepo := repositories.NewPageRepository(db)

	characterService := services.NewCharacterService(roleRepo, storageClient, aigcClient)
	comicService := services.NewComicService(comicRepo, roleRepo, sectionRepo, pageRepo, storageClient, aigcClient)
	sectionService := services.NewSectionService(comicRepo, roleRepo, sectionRepo, pageRepo, storageClient, aigcClient, characterService)
	imageService := services.NewImageService(storageClient)
	ttsService := services.NewTTSService(pageRepo, roleRepo, aigcClient)

	comicHandler := handlers.NewComicHandler(comicService)
	sectionHandler := handlers.NewSectionHandler(sectionService)
	imageHandler := handlers.NewImageHandler(imageService)
	ttsHandler := handlers.NewTTSHandler(ttsService)

	router := NewRouter(comicHandler, sectionHandler, imageHandler, ttsHandler)

	return &App{
		config: cfg,
		router: router,
	}
}

func (a *App) Run() error {
	engine := a.router.Setup()

	addr := fmt.Sprintf(":%s", a.config.Server.Port)
	fmt.Printf("Server starting on %s\n", addr)

	return engine.Run(addr)
}
