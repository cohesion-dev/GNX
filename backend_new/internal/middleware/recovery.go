package middleware

import (
	"fmt"
	"net/http"

	"github.com/cohesion-dev/GNX/backend_new/internal/utils"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic recovered: %v\n", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "server panic")
				c.Abort()
			}
		}()
		c.Next()
	}
}
