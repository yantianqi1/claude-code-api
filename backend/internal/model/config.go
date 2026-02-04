package model

import "time"

// SystemConfig represents a system configuration
type SystemConfig struct {
	ID          int64     `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ConfigCreate represents the request to create a config
type ConfigCreate struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// ConfigUpdate represents the request to update a config
type ConfigUpdate struct {
	Value       *string `json:"value"`
	Description *string `json:"description"`
}
