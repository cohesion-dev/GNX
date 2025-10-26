package app

import (
	"github.com/cohesion-dev/GNX/backend_new/internal/handlers"
	"github.com/cohesion-dev/GNX/backend_new/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine         *gin.Engine
	comicHandler   *handlers.ComicHandler
	sectionHandler *handlers.SectionHandler
	imageHandler   *handlers.ImageHandler
	ttsHandler     *handlers.TTSHandler
}

func NewRouter(
	comicHandler *handlers.ComicHandler,
	sectionHandler *handlers.SectionHandler,
	imageHandler *handlers.ImageHandler,
	ttsHandler *handlers.TTSHandler,
) *Router {
	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	return &Router{
		engine:         engine,
		comicHandler:   comicHandler,
		sectionHandler: sectionHandler,
		imageHandler:   imageHandler,
		ttsHandler:     ttsHandler,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.GET("/api/comics/", r.comicHandler.ListComics)
	r.engine.POST("/api/comics/", r.comicHandler.CreateComic)
	r.engine.GET("/api/comics/:comic_id/", r.comicHandler.GetComicDetail)

	r.engine.POST("/api/comics/:comic_id/sections/", r.sectionHandler.CreateSection)
	r.engine.GET("/api/comics/:comic_id/sections/:section_id/", r.sectionHandler.GetSectionDetail)

	r.engine.GET("/api/images/:image_id/url", r.imageHandler.GetImageURL)

	r.engine.GET("/api/tts/:tts_id", r.ttsHandler.GetTTSAudio)

	return r.engine
}
