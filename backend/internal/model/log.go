package model

import "time"

// RequestLog represents an API request log
type RequestLog struct {
	ID             int64      `json:"id"`
	ChannelID      int64      `json:"channel_id"`
	RequestID      string     `json:"request_id"`
	ModelName      string     `json:"model_name"`
	UpstreamModel  string     `json:"upstream_model"`
	InputTokens    int        `json:"input_tokens"`
	OutputTokens   int        `json:"output_tokens"`
	TotalTokens    int        `json:"total_tokens"`
	RequestTime    time.Time  `json:"request_time"`
	ResponseTime   *time.Time `json:"response_time"`
	LatencyMs      int        `json:"latency_ms"`
	Status         string     `json:"status"`
	ErrorCode      string     `json:"error_code"`
	ErrorMessage   string     `json:"error_message"`
	IPAddress      string     `json:"ip_address"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RequestLogWithChannel represents a log with channel info
type RequestLogWithChannel struct {
	ID             int64      `json:"id"`
	ChannelID      int64      `json:"channel_id"`
	ChannelName    string     `json:"channel_name"`
	RequestID      string     `json:"request_id"`
	ModelName      string     `json:"model_name"`
	UpstreamModel  string     `json:"upstream_model"`
	InputTokens    int        `json:"input_tokens"`
	OutputTokens   int        `json:"output_tokens"`
	TotalTokens    int        `json:"total_tokens"`
	RequestTime    time.Time  `json:"request_time"`
	ResponseTime   *time.Time `json:"response_time"`
	LatencyMs      int        `json:"latency_ms"`
	Status         string     `json:"status"`
	ErrorCode      string     `json:"error_code"`
	ErrorMessage   string     `json:"error_message"`
	IPAddress      string     `json:"ip_address"`
	CreatedAt      time.Time  `json:"created_at"`
}

// StatsFilter represents filter parameters for statistics
type StatsFilter struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	ChannelID int64  `form:"channel_id"`
	ModelName string `form:"model_name"`
	Status    string `form:"status"`
}

// ChannelStats represents statistics for a channel
type ChannelStats struct {
	ChannelID      int64   `json:"channel_id"`
	ChannelName    string  `json:"channel_name"`
	TotalRequests  int64   `json:"total_requests"`
	SuccessRequests int64  `json:"success_requests"`
	FailedRequests int64   `json:"failed_requests"`
	InputTokens    int64   `json:"input_tokens"`
	OutputTokens   int64   `json:"output_tokens"`
	TotalTokens    int64   `json:"total_tokens"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
}

// DailyStats represents daily statistics
type DailyStats struct {
	Date           string `json:"date"`
	TotalRequests  int64  `json:"total_requests"`
	InputTokens    int64  `json:"input_tokens"`
	OutputTokens   int64  `json:"output_tokens"`
	TotalTokens    int64  `json:"total_tokens"`
}

// ModelStats represents statistics per model
type ModelStats struct {
	ModelName      string `json:"model_name"`
	TotalRequests  int64  `json:"total_requests"`
	InputTokens    int64  `json:"input_tokens"`
	OutputTokens   int64  `json:"output_tokens"`
	TotalTokens    int64  `json:"total_tokens"`
}

// OverallStats represents overall statistics
type OverallStats struct {
	TotalChannels   int64          `json:"total_channels"`
	ActiveChannels  int64          `json:"active_channels"`
	TotalRequests   int64          `json:"total_requests"`
	TotalTokens     int64          `json:"total_tokens"`
	ChannelStats    []ChannelStats `json:"channel_stats"`
	DailyStats      []DailyStats   `json:"daily_stats"`
	ModelStats      []ModelStats   `json:"model_stats"`
}
