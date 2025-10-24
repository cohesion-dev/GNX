package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SectionHandler struct {
	db *gorm.DB
}

func NewSectionHandler(db *gorm.DB) *SectionHandler {
	return &SectionHandler{db: db}
}

func (h *SectionHandler) CreateSection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create section"})
}

func (h *SectionHandler) GetSectionContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get section content"})
}

func (h *SectionHandler) GetStoryboards(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get storyboards"})
}
