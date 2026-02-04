package model

import "time"

// ModelMapping represents a model name mapping
type ModelMapping struct {
	ID             int64     `json:"id"`
	ChannelID      int64     `json:"channel_id"`
	UpstreamModel  string    `json:"upstream_model"`
	DisplayModel   string    `json:"display_model"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MappingCreate represents the request to create a mapping
type MappingCreate struct {
	ChannelID     int64  `json:"channel_id" binding:"required"`
	UpstreamModel string `json:"upstream_model" binding:"required"`
	DisplayModel  string `json:"display_model" binding:"required"`
}

// MappingUpdate represents the request to update a mapping
type MappingUpdate struct {
	UpstreamModel *string `json:"upstream_model"`
	DisplayModel  *string `json:"display_model"`
	IsEnabled     *bool   `json:"is_enabled"`
}

// ModelMappingWithChannel represents a mapping with channel info
type ModelMappingWithChannel struct {
	ID             int64     `json:"id"`
	ChannelID      int64     `json:"channel_id"`
	ChannelName    string    `json:"channel_name"`
	UpstreamModel  string    `json:"upstream_model"`
	DisplayModel   string    `json:"display_model"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
