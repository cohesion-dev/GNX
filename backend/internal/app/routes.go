package app

import (
	"github.com/cohesion-dev/GNX/backend/internal/handlers"
)

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		comic := api.Group("/comic")
		{
			comic.GET("", handlers.GetComics)
			comic.POST("", handlers.CreateComic)
			comic.GET("/:id", handlers.GetComic)
			comic.GET("/:id/sections", handlers.GetComicSections)

			comic.POST("/:id/section", handlers.CreateSection)
			comic.GET("/:id/section/:section_id/content", handlers.GetSectionContent)
			comic.GET("/:id/section/:section_id/storyboards", handlers.GetStoryboards)
		}

		api.GET("/tts/:storyboard_tts_id", handlers.GetTTSAudio)
	}
}
