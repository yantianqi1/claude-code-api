package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/pkg/logger"
)

// Recovery is a panic recovery middleware
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
