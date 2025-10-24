package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TTSHandler struct {
	db *gorm.DB
}

func NewTTSHandler(db *gorm.DB) *TTSHandler {
	return &TTSHandler{db: db}
}

func (h *TTSHandler) GetTTSAudio(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get TTS audio"})
}
