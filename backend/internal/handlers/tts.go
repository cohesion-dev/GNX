package handlers

import (
	"github.com/gin-gonic/gin"
)

func GetTTSAudio(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get TTS audio"})
}
