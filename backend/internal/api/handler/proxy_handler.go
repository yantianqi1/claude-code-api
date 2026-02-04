package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/proxy"
)

// ProxyHandler handles proxy API requests
type ProxyHandler struct {
	proxyService *proxy.ProxyService
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxy.NewProxyService(),
	}
}

// ProxyMessage handles the /v1/messages endpoint
func (h *ProxyHandler) ProxyMessage(c *gin.Context) {
	// Get API key from header
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		// Try Authorization header
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Parse request body
	var req model.AnthropicMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "invalid_request_error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	// Get client IP
	ipAddress := c.ClientIP()

	// Handle streaming
	if req.Stream {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		if err := h.proxyService.ProxyMessageStream(&req, apiKey, ipAddress, c.Writer); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"type": "error",
				"error": gin.H{
					"type":    "api_error",
					"message": err.Error(),
				},
			})
		}
		return
	}

	// Handle non-streaming request
	resp, err := h.proxyService.ProxyMessage(&req, apiKey, ipAddress)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// StreamHandler handles SSE streaming
func (h *ProxyHandler) StreamHandler(c *gin.Context) {
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	var req model.AnthropicMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Stream = true
	ipAddress := c.ClientIP()

	if err := h.proxyService.ProxyMessageStream(&req, apiKey, ipAddress, c.Writer); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	}
}

// HealthCheck returns the health status
func (h *ProxyHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "claude-api-gateway",
	})
}

// ProxyGet handles GET requests (for compatibility)
func (h *ProxyHandler) ProxyGet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Claude API Gateway",
		"version": "1.0.0",
		"endpoints": []string{
			"POST /v1/messages - Anthropic Messages API",
			"GET  /api/health - Health check",
			"GET  /api/channels - List channels",
			"POST /api/channels - Create channel",
			"GET  /api/stats - Get statistics",
		},
	})
}
