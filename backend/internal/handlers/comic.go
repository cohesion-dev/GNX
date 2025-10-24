package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ComicHandler struct {
	db *gorm.DB
}

func NewComicHandler(db *gorm.DB) *ComicHandler {
	return &ComicHandler{db: db}
}

func (h *ComicHandler) GetComics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get comics"})
}

func (h *ComicHandler) CreateComic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create comic"})
}

func (h *ComicHandler) GetComic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get comic"})
}

func (h *ComicHandler) GetComicSections(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get comic sections"})
}
