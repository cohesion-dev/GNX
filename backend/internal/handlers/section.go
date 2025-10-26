package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cohesion-dev/GNX/backend/internal/services"
	"github.com/cohesion-dev/GNX/backend/internal/utils"
)

type SectionHandler struct {
	sectionService *services.SectionService
}

func NewSectionHandler(sectionService *services.SectionService) *SectionHandler {
	return &SectionHandler{
		sectionService: sectionService,
	}
}

func (h *SectionHandler) CreateSection(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	var req struct {
		Index  int    `json:"index" binding:"required"`
		Detail string `json:"detail" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	section, err := h.sectionService.CreateSection(uint(comicID), req.Index, req.Detail)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create section", err.Error())
		return
	}

	utils.SuccessResponseWithStatus(c, http.StatusCreated, section)
}

func (h *SectionHandler) GetSectionContent(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	sectionID, err := strconv.ParseUint(c.Param("section_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid section ID", err.Error())
		return
	}

	content, err := h.sectionService.GetSectionContent(uint(comicID), uint(sectionID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Section not found", err.Error())
		return
	}

	utils.SuccessResponse(c, content)
}

func (h *SectionHandler) GetStoryboards(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	sectionID, err := strconv.ParseUint(c.Param("section_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid section ID", err.Error())
		return
	}

	storyboards, err := h.sectionService.GetStoryboards(uint(comicID), uint(sectionID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get storyboards", err.Error())
		return
	}

	utils.SuccessResponse(c, storyboards)
}

func (h *SectionHandler) GetStoryboardImage(c *gin.Context) {
	panelID, err := strconv.ParseUint(c.Param("panel_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid panel ID", err.Error())
		return
	}

	imageData, err := h.sectionService.GetStoryboardImage(uint(panelID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Image not found", err.Error())
		return
	}

	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", strconv.Itoa(len(imageData)))
	c.Data(http.StatusOK, "image/png", imageData)
}
