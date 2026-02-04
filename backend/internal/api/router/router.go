package router

import (
	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/internal/api/handler"
	"github.com/claude-api-gateway/backend/internal/api/middleware"
	"github.com/claude-api-gateway/backend/internal/config"
)

// Setup configures all routes
func Setup(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(middleware.Recovery())
	if cfg.EnableCORS {
		r.Use(middleware.CORS(cfg.AllowedOrigins))
	}
	r.Use(middleware.RequestLogger())

	// Initialize handlers
	proxyHandler := handler.NewProxyHandler(cfg.APIKey)
	authHandler := handler.NewAuthHandler(cfg.APIKey)
	channelHandler := handler.NewChannelHandler()
	mappingHandler := handler.NewMappingHandler()
	statsHandler := handler.NewStatsHandler()

	// Root and health
	r.GET("/", proxyHandler.ProxyGet)
	r.GET("/api/health", proxyHandler.HealthCheck)

	// Auth API (no middleware required)
	api := r.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)
		api.GET("/auth/verify", middleware.AuthMiddleware(cfg.APIKey), authHandler.Verify)
	}

	// Anthropic API compatible endpoint
	r.POST("/v1/messages", proxyHandler.ProxyMessage)

	// Management API (protected by auth middleware)
	managementAPI := r.Group("/api")
	managementAPI.Use(middleware.AuthMiddleware(cfg.APIKey))
	{
		// Channels
		managementAPI.GET("/channels", channelHandler.List)
		managementAPI.POST("/channels", channelHandler.Create)
		managementAPI.GET("/channels/:id", channelHandler.Get)
		managementAPI.PUT("/channels/:id", channelHandler.Update)
		managementAPI.DELETE("/channels/:id", channelHandler.Delete)
		managementAPI.PUT("/channels/:id/activate", channelHandler.Activate)
		managementAPI.PUT("/channels/:id/deactivate", channelHandler.Deactivate)
		managementAPI.POST("/channels/test", channelHandler.Test)
		managementAPI.GET("/channels/:id/mappings", mappingHandler.ListByChannel)

		// Model Mappings
		managementAPI.GET("/mappings", mappingHandler.List)
		managementAPI.POST("/mappings", mappingHandler.Create)
		managementAPI.GET("/mappings/:id", mappingHandler.Get)
		managementAPI.PUT("/mappings/:id", mappingHandler.Update)
		managementAPI.DELETE("/mappings/:id", mappingHandler.Delete)

		// Statistics
		managementAPI.GET("/stats", statsHandler.GetOverall)
		managementAPI.GET("/stats/channels", statsHandler.GetChannelStats)
		managementAPI.GET("/stats/daily", statsHandler.GetDailyStats)
		managementAPI.GET("/stats/models", statsHandler.GetModelStats)
		managementAPI.GET("/stats/logs", statsHandler.GetLogs)
		managementAPI.GET("/stats/export", statsHandler.Export)
	}

	return r
}
