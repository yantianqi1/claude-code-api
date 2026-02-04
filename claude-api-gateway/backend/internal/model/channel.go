package model

import "time"

// Channel represents an upstream API channel
type Channel struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	BaseURL    string     `json:"base_url"`
	APIKey     string     `json:"api_key"`
	Provider   string     `json:"provider"`
	IsActive   bool       `json:"is_active"`
	Priority   int        `json:"priority"`
	MaxRetries int        `json:"max_retries"`
	Timeout    int        `json:"timeout"`
	RateLimit  int        `json:"rate_limit"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ChannelCreate represents the request to create a channel
type ChannelCreate struct {
	Name       string `json:"name" binding:"required"`
	BaseURL    string `json:"base_url" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
	Provider   string `json:"provider" binding:"required"`
	Priority   int    `json:"priority"`
	MaxRetries int    `json:"max_retries"`
	Timeout    int    `json:"timeout"`
	RateLimit  int    `json:"rate_limit"`
}

// ChannelUpdate represents the request to update a channel
type ChannelUpdate struct {
	Name       *string `json:"name"`
	BaseURL    *string `json:"base_url"`
	APIKey     *string `json:"api_key"`
	Provider   *string `json:"provider"`
	IsActive   *bool   `json:"is_active"`
	Priority   *int    `json:"priority"`
	MaxRetries *int    `json:"max_retries"`
	Timeout    *int    `json:"timeout"`
	RateLimit  *int    `json:"rate_limit"`
}

// ChannelTestRequest represents the request to test a channel
type ChannelTestRequest struct {
	BaseURL string `json:"base_url" binding:"required"`
	APIKey  string `json:"api_key" binding:"required"`
}
