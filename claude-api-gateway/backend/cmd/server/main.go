package main

import (
	"fmt"
	"log"
	"os"

	"github.com/claude-api-gateway/backend/internal/api/router"
	"github.com/claude-api-gateway/backend/internal/config"
	"github.com/claude-api-gateway/backend/pkg/database"
	"github.com/claude-api-gateway/backend/pkg/logger"
)

func runMigrations() error {
	// Read migration file from filesystem
	migrationPath := "migrations/001_init.sql"
	migrations, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	_, err = database.DB.Exec(string(migrations))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Initialize(cfg.Debug)

	logger.Info("Starting Claude API Gateway...")
	logger.Info("Configuration: Port=%d, DataDir=%s, Debug=%v", cfg.ServerPort, cfg.DataDir, cfg.Debug)

	// Initialize database
	if err := database.Initialize(cfg.DataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	logger.Info("Database initialized at %s/gateway.db", cfg.DataDir)

	// Run migrations
	if err := runMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup router
	r := router.Setup(cfg.EnableCORS, cfg.AllowedOrigins)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	logger.Info("Server starting on %s", addr)
	logger.Info("API endpoints available:")
	logger.Info("  - POST /v1/messages (Anthropic API)")
	logger.Info("  - GET  /api/channels")
	logger.Info("  - POST /api/channels")
	logger.Info("  - GET  /api/stats")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
