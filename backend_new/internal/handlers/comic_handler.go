package handlers

import (
	"net/http"
	"strconv"

	"github.com/cohesion-dev/GNX/backend_new/internal/services"
	"github.com/cohesion-dev/GNX/backend_new/internal/utils"
	"github.com/gin-gonic/gin"
)

type ComicHandler struct {
	comicService *services.ComicService
}

func NewComicHandler(comicService *services.ComicService) *ComicHandler {
	return &ComicHandler{comicService: comicService}
}

func (h *ComicHandler) ListComics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	comics, total, err := h.comicService.GetComicList(page, limit, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
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
	title := c.PostForm("title")
	userPrompt := c.PostForm("user_prompt")

	if title == "" || userPrompt == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "title and user_prompt are required")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "file is required")
		return
	}

	fileContent, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	defer fileContent.Close()

	comic, err := h.comicService.CreateComic(c.Request.Context(), title, userPrompt, fileContent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"id": strconv.FormatUint(uint64(comic.ID), 10),
	})
}

func (h *ComicHandler) GetComicDetail(c *gin.Context) {
	comicID, err := strconv.ParseUint(c.Param("comic_id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "invalid comic_id")
		return
	}

	comic, err := h.comicService.GetComicDetail(uint(comicID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", err.Error())
		return
	}

	utils.SuccessResponse(c, comic)
}
