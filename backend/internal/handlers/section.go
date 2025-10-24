package handlers

import (
	"github.com/gin-gonic/gin"
)

func CreateSection(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create section"})
}

func GetSectionContent(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get section content"})
}

func GetStoryboards(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get storyboards"})
}
