package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cohesion-dev/GNX/backend/internal/services"
	"github.com/cohesion-dev/GNX/backend/internal/utils"
)

type TTSHandler struct {
	ttsService *services.TTSService
}

func NewTTSHandler(ttsService *services.TTSService) *TTSHandler {
	return &TTSHandler{
		ttsService: ttsService,
	}
}

func (h *TTSHandler) GetTTSAudio(c *gin.Context) {
	detailID, err := strconv.ParseUint(c.Param("storyboard_tts_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid TTS ID", err.Error())
		return
	}

	audioData, err := h.ttsService.GetTTSAudio(uint(detailID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "TTS audio not found", err.Error())
		return
	}

	c.Data(http.StatusOK, "audio/mpeg", audioData)
}
