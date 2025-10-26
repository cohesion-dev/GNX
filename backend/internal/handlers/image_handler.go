package handlers

import (
	"net/http"

	"github.com/cohesion-dev/GNX/backend_new/internal/services"
	"github.com/cohesion-dev/GNX/backend_new/internal/utils"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService
}

func NewImageHandler(imageService *services.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

func (h *ImageHandler) GetImageURL(c *gin.Context) {
	imageID := c.Param("image_id")

	url, err := h.imageService.GetImageURL(imageID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"url": url,
	})
}
