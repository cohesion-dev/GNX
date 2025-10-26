package handlers

import (
	"net/http"
	"strconv"

	"github.com/cohesion-dev/GNX/backend_new/internal/services"
	"github.com/cohesion-dev/GNX/backend_new/internal/utils"
	"github.com/gin-gonic/gin"
)

type TTSHandler struct {
	ttsService *services.TTSService
}

func NewTTSHandler(ttsService *services.TTSService) *TTSHandler {
	return &TTSHandler{ttsService: ttsService}
}

func (h *TTSHandler) GetTTSAudio(c *gin.Context) {
	ttsID, err := strconv.ParseUint(c.Param("tts_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "invalid tts_id")
		return
	}

	audioData, err := h.ttsService.GetTTSAudio(c.Request.Context(), uint(ttsID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", err.Error())
		return
	}

	c.Header("Content-Type", "audio/mpeg")
	c.Header("Content-Length", strconv.Itoa(len(audioData)))
	c.Data(http.StatusOK, "audio/mpeg", audioData)
}
