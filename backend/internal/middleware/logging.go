package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)

		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		fmt.Printf("[%s] %s %s %d %v\n",
			startTime.Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			latency,
		)
	}
}
