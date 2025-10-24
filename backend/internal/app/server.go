package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/config"
	"github.com/cohesion-dev/GNX/backend/internal/middleware"
)

type Server struct {
	db     *gorm.DB
	config *Config
	router *gin.Engine
}

func NewServer(db *gorm.DB, cfg *config.Config) *Server {
	router := gin.Default()

	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())
	router.Use(middleware.Logging())

	server := &Server{
		db:     db,
		config: cfg,
		router: router,
	}

	server.setupRoutes()

	return server
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.config.ServerPort)
	return s.router.Run(addr)
}
