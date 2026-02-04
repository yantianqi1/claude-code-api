package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication
type AuthHandler struct {
	apiKey string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(apiKey string) *AuthHandler {
	return &AuthHandler{
		apiKey: apiKey,
	}
}

// LoginRequest represents the login request
type LoginRequest struct {
	APIKey string `json:"api_key" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// Login handles login requests
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// Validate API key
	if req.APIKey != h.apiKey {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "Invalid API key",
		})
		return
	}

	// Set cookie for web UI
	c.SetCookie(
		"auth_token",
		h.apiKey,
		3600*24*30, // 30 days
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   h.apiKey,
	})
}

// Logout handles logout requests
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

// Verify verifies the current authentication status
func (h *AuthHandler) Verify(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
	})
}
