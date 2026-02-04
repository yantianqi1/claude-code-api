package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the API key for management endpoints
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if no API key is configured
		if apiKey == "" {
			c.Next()
			return
		}

		// Get API key from header
		clientKey := c.GetHeader("x-api-key")
		if clientKey == "" {
			// Try Authorization header
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				clientKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		// Try cookie (for web UI)
		if clientKey == "" {
			if cookie, err := c.Cookie("auth_token"); err == nil {
				clientKey = cookie
			}
		}

		if clientKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
