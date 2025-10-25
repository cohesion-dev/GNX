package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/handlers"
	"github.com/cohesion-dev/GNX/backend/internal/middleware"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/internal/services"
	"github.com/cohesion-dev/GNX/backend/pkg/ai"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type Server struct {
	db             *gorm.DB
	config         *config.Config
	router         *gin.Engine
	comicHandler   *handlers.ComicHandler
	sectionHandler *handlers.SectionHandler
	ttsHandler     *handlers.TTSHandler
}

func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	router := gin.Default()

	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())
	router.Use(middleware.Logging())

	comicRepo := repositories.NewComicRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	sectionRepo := repositories.NewSectionRepository(db)
	storyboardRepo := repositories.NewStoryboardRepository(db)

	aiService := ai.NewOpenAIClient(cfg.OpenAI.APIKey)
	aigcService := gnxaigc.NewGnxAIGC(gnxaigc.Config{
		APIKey:        cfg.OpenAI.APIKey,
		BaseURL:       cfg.OpenAI.BaseURL,
		ImageModel:    cfg.OpenAI.ImageModel,
		LanguageModel: cfg.OpenAI.LanguageModel,
	})
	storageService := storage.NewQiniuClient(
		cfg.Qiniu.AccessKey,
		cfg.Qiniu.SecretKey,
		cfg.Qiniu.Bucket,
		cfg.Qiniu.Domain,
	)

	comicService := services.NewComicService(
		comicRepo,
		roleRepo,
		sectionRepo,
		storyboardRepo,
		storageService,
		aigcService,
		db,
	)

	sectionService := services.NewSectionService(
		sectionRepo,
		storyboardRepo,
		roleRepo,
		comicRepo,
		aiService,
		storageService,
		db,
	)

	ttsService := services.NewTTSService(storyboardRepo, aigcService, storageService)

	comicHandler := handlers.NewComicHandler(comicService)
	sectionHandler := handlers.NewSectionHandler(sectionService)
	ttsHandler := handlers.NewTTSHandler(ttsService)

	server := &Server{
		db:             db,
		config:         cfg,
		router:         router,
		comicHandler:   comicHandler,
		sectionHandler: sectionHandler,
		ttsHandler:     ttsHandler,
	}

	server.setupRoutes()

	return server
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	return s.router.Run(addr)
}
