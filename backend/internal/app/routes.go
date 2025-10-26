package app

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		comic := api.Group("/comic")
		{
			comic.GET("", s.comicHandler.GetComics)
			comic.POST("", s.comicHandler.CreateComic)
			comic.GET("/:id", s.comicHandler.GetComic)
			comic.GET("/:id/sections", s.comicHandler.GetComicSections)
			comic.POST("/:id/sections", s.comicHandler.AppendSections)

			comic.POST("/:id/section", s.sectionHandler.CreateSection)
			comic.GET("/:id/section/:section_id/content", s.sectionHandler.GetSectionContent)
			comic.GET("/:id/section/:section_id/storyboards", s.sectionHandler.GetStoryboards)
		}

		api.GET("/panel/:panel_id/image", s.sectionHandler.GetStoryboardImage)
		api.GET("/tts/:storyboard_tts_id", s.ttsHandler.GetTTSAudio)
	}
}
