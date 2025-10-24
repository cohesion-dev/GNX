package app

import (
	"github.com/cohesion-dev/GNX/backend/internal/handlers"
	"github.com/cohesion-dev/GNX/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func (s *Server) SetupRoutes(r *gin.Engine) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logging())
	r.Use(middleware.Recovery())

	api := r.Group("/api")
	{
		comicHandler := handlers.NewComicHandler(s.DB)
		sectionHandler := handlers.NewSectionHandler(s.DB)
		ttsHandler := handlers.NewTTSHandler(s.DB)

		api.GET("/comic", comicHandler.GetComics)
		api.POST("/comic", comicHandler.CreateComic)
		api.GET("/comic/:id", comicHandler.GetComic)
		api.GET("/comic/:id/sections", comicHandler.GetComicSections)

		api.POST("/comic/:id/section", sectionHandler.CreateSection)
		api.GET("/comic/:id/section/:section_id/content", sectionHandler.GetSectionContent)
		api.GET("/comic/:id/section/:section_id/storyboards", sectionHandler.GetStoryboards)

		api.GET("/tts/:storyboard_tts_id", ttsHandler.GetTTSAudio)
	}
}
