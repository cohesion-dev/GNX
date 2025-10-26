package handlers

import (
	"net/http"
	"strconv"

	"github.com/cohesion-dev/GNX/backend_new/internal/services"
	"github.com/cohesion-dev/GNX/backend_new/internal/utils"
	"github.com/gin-gonic/gin"
)

type SectionHandler struct {
	sectionService *services.SectionService
}

func NewSectionHandler(sectionService *services.SectionService) *SectionHandler {
	return &SectionHandler{sectionService: sectionService}
}

func (h *SectionHandler) CreateSection(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("comic_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "invalid comic_id")
		return
	}

	title := c.PostForm("title")
	content := c.PostForm("content")

	if title == "" || content == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "title and content are required")
		return
	}

	section, err := h.sectionService.CreateSection(c.Request.Context(), uint(comicID), title, content)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"id":    strconv.FormatUint(uint64(section.ID), 10),
		"index": section.Index,
	})
}

func (h *SectionHandler) GetSectionDetail(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("comic_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "invalid comic_id")
		return
	}

	sectionID, err := strconv.ParseUint(c.Param("section_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "invalid section_id")
		return
	}

	section, err := h.sectionService.GetSectionDetail(uint(comicID), uint(sectionID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", err.Error())
		return
	}

	utils.SuccessResponse(c, section)
}
