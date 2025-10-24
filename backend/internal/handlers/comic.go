package handlers

import (
	"github.com/gin-gonic/gin"
)

func GetComics(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get comics"})
}

func CreateComic(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create comic"})
}

func GetComic(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get comic"})
}

func GetComicSections(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get comic sections"})
}
