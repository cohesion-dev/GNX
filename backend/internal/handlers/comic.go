package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cohesion-dev/GNX/backend/internal/services"
	"github.com/cohesion-dev/GNX/backend/internal/utils"
)

type ComicHandler struct {
	comicService *services.ComicService
}

func NewComicHandler(comicService *services.ComicService) *ComicHandler {
	return &ComicHandler{
		comicService: comicService,
	}
}

func (h *ComicHandler) GetComics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 100
	}

	comics, total, err := h.comicService.GetComicList(page, limit, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get comics", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"comics": comics,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *ComicHandler) CreateComic(c *gin.Context) {
	var req struct {
		Title      string `form:"title" binding:"required"`
		UserPrompt string `form:"user_prompt" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required", err.Error())
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read file", err.Error())
		return
	}
	defer fileContent.Close()

	comic, err := h.comicService.CreateComic(req.Title, req.UserPrompt, fileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create comic", err.Error())
		return
	}

	utils.SuccessResponseWithStatus(c, http.StatusCreated, comic)
}

func (h *ComicHandler) GetComic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	comic, err := h.comicService.GetComicDetail(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Comic not found", err.Error())
		return
	}

	utils.SuccessResponse(c, comic)
}

func (h *ComicHandler) GetComicSections(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	sections, err := h.comicService.GetComicSections(uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get sections", err.Error())
		return
	}

	utils.SuccessResponse(c, sections)
}

func (h *ComicHandler) AppendSections(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid comic ID", err.Error())
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required", err.Error())
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to read file", err.Error())
		return
	}
	defer fileContent.Close()

	if err := h.comicService.AppendSections(uint(id), fileContent); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to append sections", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Sections appending started"})
}
