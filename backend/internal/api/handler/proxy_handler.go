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
	apiKey       string
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(apiKey string) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxy.NewProxyService(),
		apiKey:       apiKey,
	}
}

// validateAPIKey validates the API key from request
func (h *ProxyHandler) validateAPIKey(c *gin.Context) (string, bool) {
	// Get API key from header
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		// Try Authorization header
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Try cookie (for web UI)
	if apiKey == "" {
		if cookie, err := c.Cookie("auth_token"); err == nil {
			apiKey = cookie
		}
	}

	// If no API key configured, allow access
	if h.apiKey == "" {
		return apiKey, true
	}

	// Validate API key
	if apiKey != h.apiKey {
		return "", false
	}

	return apiKey, true
}

// ProxyMessage handles the /v1/messages endpoint
func (h *ProxyHandler) ProxyMessage(c *gin.Context) {
	// Validate API key
	apiKey, valid := h.validateAPIKey(c)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "authentication_error",
				"message": "Invalid API key",
			},
		})
		return
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
	// Validate API key
	apiKey, valid := h.validateAPIKey(c)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "authentication_error",
				"message": "Invalid API key",
			},
		})
		return
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
			"POST /v1/chat/completions - OpenAI Chat Completions API",
			"GET  /api/health - Health check",
			"GET  /api/channels - List channels",
			"POST /api/channels - Create channel",
			"GET  /api/stats - Get statistics",
		},
	})
}

// ProxyChatCompletions handles the /v1/chat/completions endpoint (OpenAI compatible)
func (h *ProxyHandler) ProxyChatCompletions(c *gin.Context) {
	// Validate API key
	apiKey, valid := h.validateAPIKey(c)
	if !valid {
		c.JSON(http.StatusUnauthorized, model.OpenAIErrorResponse{
			Error: model.OpenAIErrorDetail{
				Type:    "invalid_request_error",
				Message: "Incorrect API key provided",
			},
		})
		return
	}

	// Parse request body
	var req model.OpenAIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.OpenAIErrorResponse{
			Error: model.OpenAIErrorDetail{
				Type:    "invalid_request_error",
				Message: "Invalid request body: " + err.Error(),
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

		if err := h.proxyService.ProxyChatStream(&req, apiKey, ipAddress, c.Writer); err != nil {
			c.JSON(http.StatusBadGateway, model.OpenAIErrorResponse{
				Error: model.OpenAIErrorDetail{
					Type:    "api_error",
					Message: err.Error(),
				},
			})
		}
		return
	}

	// Handle non-streaming request
	resp, err := h.proxyService.ProxyChat(&req, apiKey, ipAddress)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.OpenAIErrorResponse{
			Error: model.OpenAIErrorDetail{
				Type:    "api_error",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListModels handles GET /v1/models endpoint
func (h *ProxyHandler) ListModels(c *gin.Context) {
	// Return a static list of available models
	// In production, you might want to fetch this from database or upstream
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"id":      "claude-sonnet-4-5-thinking",
				"object":  "model",
				"created": 1234567890,
				"owned_by": "anthropic",
			},
			{
				"id":      "claude-opus-4-5-thinking",
				"object":  "model",
				"created": 1234567890,
				"owned_by": "anthropic",
			},
			{
				"id":      "claude-haiku-4-5-20251001",
				"object":  "model",
				"created": 1234567890,
				"owned_by": "anthropic",
			},
		},
	})
}
