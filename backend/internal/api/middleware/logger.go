package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/pkg/logger"
)

// RequestLogger is a request logging middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		logger.Info("%s %s %s - Status: %d - Latency: %v",
			c.Request.Method,
			path,
			query,
			c.Writer.Status(),
			latency,
		)
	}
}
