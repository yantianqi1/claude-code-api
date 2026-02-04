package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	ServerPort     int
	DataDir        string
	APIKey         string
	Debug          bool
	EnableCORS     bool
	AllowedOrigins string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		ServerPort:     getEnvInt("SERVER_PORT", 8080),
		DataDir:        getEnv("DATA_DIR", "./data"),
		APIKey:         getEnv("API_KEY", ""),
		Debug:          getEnvBool("DEBUG", false),
		EnableCORS:     getEnvBool("ENABLE_CORS", true),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
