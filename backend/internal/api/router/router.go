package router

import (
	"github.com/gin-gonic/gin"
	"github.com/claude-api-gateway/backend/internal/api/handler"
	"github.com/claude-api-gateway/backend/internal/api/middleware"
)

// Setup configures all routes
func Setup(enableCORS bool, allowedOrigins string) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(middleware.Recovery())
	if enableCORS {
		r.Use(middleware.CORS(allowedOrigins))
	}
	r.Use(middleware.RequestLogger())

	// Initialize handlers
	proxyHandler := handler.NewProxyHandler()
	channelHandler := handler.NewChannelHandler()
	mappingHandler := handler.NewMappingHandler()
	statsHandler := handler.NewStatsHandler()

	// Root and health
	r.GET("/", proxyHandler.ProxyGet)
	r.GET("/api/health", proxyHandler.HealthCheck)

	// Anthropic API compatible endpoint
	r.POST("/v1/messages", proxyHandler.ProxyMessage)

	// Management API
	api := r.Group("/api")
	{
		// Channels
		api.GET("/channels", channelHandler.List)
		api.POST("/channels", channelHandler.Create)
		api.GET("/channels/:id", channelHandler.Get)
		api.PUT("/channels/:id", channelHandler.Update)
		api.DELETE("/channels/:id", channelHandler.Delete)
		api.PUT("/channels/:id/activate", channelHandler.Activate)
		api.PUT("/channels/:id/deactivate", channelHandler.Deactivate)
		api.POST("/channels/test", channelHandler.Test)
		api.GET("/channels/:id/mappings", mappingHandler.ListByChannel)

		// Model Mappings
		api.GET("/mappings", mappingHandler.List)
		api.POST("/mappings", mappingHandler.Create)
		api.GET("/mappings/:id", mappingHandler.Get)
		api.PUT("/mappings/:id", mappingHandler.Update)
		api.DELETE("/mappings/:id", mappingHandler.Delete)

		// Statistics
		api.GET("/stats", statsHandler.GetOverall)
		api.GET("/stats/channels", statsHandler.GetChannelStats)
		api.GET("/stats/daily", statsHandler.GetDailyStats)
		api.GET("/stats/models", statsHandler.GetModelStats)
		api.GET("/stats/logs", statsHandler.GetLogs)
		api.GET("/stats/export", statsHandler.Export)
	}

	return r
}
