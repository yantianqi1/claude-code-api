package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/claude-api-gateway/backend/internal/model"
	"github.com/claude-api-gateway/backend/internal/repository"
	"github.com/claude-api-gateway/backend/pkg/logger"
	"github.com/google/uuid"
)

// ProxyService handles API proxy operations
type ProxyService struct {
	channelRepo *repository.ChannelRepository
	mappingRepo *repository.MappingRepository
	logRepo     *repository.LogRepository
	client      *http.Client
}

// NewProxyService creates a new proxy service
func NewProxyService() *ProxyService {
	return &ProxyService{
		channelRepo: repository.NewChannelRepository(),
		mappingRepo: repository.NewMappingRepository(),
		logRepo:     repository.NewLogRepository(),
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ProxyMessage proxies a request to the Anthropic Messages API
func (s *ProxyService) ProxyMessage(req *model.AnthropicMessageRequest, apiKey string, ipAddress string) (*model.AnthropicMessageResponse, error) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Find model mapping
	mappings, err := s.mappingRepo.FindByDisplayModel(req.Model)
	if err != nil || len(mappings) == 0 {
		// No mapping found, use the model name as is and get first active channel
		logger.Debug("No mapping found for model: %s", req.Model)
		return s.proxyToChannel(req, apiKey, ipAddress, requestID, startTime, nil, req.Model)
	}

	// Try each mapped channel in priority order
	for _, mapping := range mappings {
		if !mapping.IsEnabled {
			continue
		}

		channel, err := s.channelRepo.GetByID(mapping.ChannelID)
		if err != nil {
			logger.Error("Failed to get channel %d: %v", mapping.ChannelID, err)
			continue
		}

		if !channel.IsActive {
			logger.Debug("Channel %d is not active", channel.ID)
			continue
		}

		resp, err := s.proxyToChannel(req, apiKey, ipAddress, requestID, startTime, channel, mapping.UpstreamModel)
		if err != nil {
			logger.Error("Failed to proxy to channel %d: %v", channel.ID, err)
			continue
		}
		return resp, nil
	}

	return nil, fmt.Errorf("all channels failed for model: %s", req.Model)
}

// proxyToChannel proxies a request to a specific channel
func (s *ProxyService) proxyToChannel(req *model.AnthropicMessageRequest, apiKey string, ipAddress string, requestID string, startTime time.Time, channel *model.Channel, upstreamModel string) (*model.AnthropicMessageResponse, error) {
	var channelID int64
	var baseURL string
	var apiKeyToUse string
	var timeout time.Duration

	if channel != nil {
		channelID = channel.ID
		baseURL = channel.BaseURL
		apiKeyToUse = channel.APIKey
		timeout = time.Duration(channel.Timeout) * time.Second
	} else {
		// Use the API key from request to find channel
		channelID = 0
		baseURL = "https://api.anthropic.com"
		apiKeyToUse = apiKey
		timeout = 120 * time.Second
	}

	// Clone the request and update the model
	proxyReq := &model.AnthropicMessageRequest{
		Model:         upstreamModel,
		MaxTokens:     req.MaxTokens,
		Messages:      req.Messages,
		Temperature:   req.Temperature,
		TopP:          req.TopP,
		TopK:          req.TopK,
		Stream:        req.Stream,
		StopSequences: req.StopSequences,
		System:        req.System,
		Tools:         req.Tools,
		ToolChoice:    req.ToolChoice,
		Metadata:      req.Metadata,
	}

	// Marshal request body
	bodyBytes, err := json.Marshal(proxyReq)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKeyToUse)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Make request
	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "http_request", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "read_response", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if httpResp.StatusCode != http.StatusOK {
		var errResp model.AnthropicErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			s.logError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, errResp.Error.Type, errResp.Error.Message, ipAddress)
			return nil, fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "unknown", string(respBody), ipAddress)
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	// Parse success response
	var response model.AnthropicMessageResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "parse_response", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Log successful request
	s.logSuccess(channelID, requestID, req.Model, upstreamModel, startTime, &response, ipAddress)

	return &response, nil
}

// ProxyMessageStream proxies a streaming request
func (s *ProxyService) ProxyMessageStream(req *model.AnthropicMessageRequest, apiKey string, ipAddress string, w http.ResponseWriter) error {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Find model mapping
	mappings, err := s.mappingRepo.FindByDisplayModel(req.Model)
	if err != nil || len(mappings) == 0 {
		return s.proxyStreamToChannel(req, apiKey, ipAddress, requestID, startTime, nil, req.Model, w)
	}

	// Try each mapped channel
	for _, mapping := range mappings {
		if !mapping.IsEnabled {
			continue
		}

		channel, err := s.channelRepo.GetByID(mapping.ChannelID)
		if err != nil || !channel.IsActive {
			continue
		}

		if err := s.proxyStreamToChannel(req, apiKey, ipAddress, requestID, startTime, channel, mapping.UpstreamModel, w); err != nil {
			logger.Error("Stream proxy to channel %d failed: %v", channel.ID, err)
			continue
		}
		return nil
	}

	return fmt.Errorf("all channels failed for streaming model: %s", req.Model)
}

// proxyStreamToChannel proxies a streaming request to a specific channel
func (s *ProxyService) proxyStreamToChannel(req *model.AnthropicMessageRequest, apiKey string, ipAddress string, requestID string, startTime time.Time, channel *model.Channel, upstreamModel string, w http.ResponseWriter) error {
	var channelID int64
	var baseURL string
	var apiKeyToUse string
	var timeout time.Duration

	if channel != nil {
		channelID = channel.ID
		baseURL = channel.BaseURL
		apiKeyToUse = channel.APIKey
		timeout = time.Duration(channel.Timeout) * time.Second
	} else {
		channelID = 0
		baseURL = "https://api.anthropic.com"
		apiKeyToUse = apiKey
		timeout = 120 * time.Second
	}

	// Clone the request
	proxyReq := &model.AnthropicMessageRequest{
		Model:         upstreamModel,
		MaxTokens:     req.MaxTokens,
		Messages:      req.Messages,
		Temperature:   req.Temperature,
		TopP:          req.TopP,
		TopK:          req.TopK,
		Stream:        true,
		StopSequences: req.StopSequences,
		System:        req.System,
		Tools:         req.Tools,
		ToolChoice:    req.ToolChoice,
		Metadata:      req.Metadata,
	}

	bodyBytes, _ := json.Marshal(proxyReq)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKeyToUse)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "http_request", err.Error(), ipAddress)
		return err
	}
	defer httpResp.Body.Close()

	// Set streaming headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Copy stream
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	inputTokens := 0
	outputTokens := 0

	decoder := json.NewDecoder(httpResp.Body)
	for {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		// Track token usage
		if eventType, ok := event["type"].(string); ok {
			if eventType == "message_stop" {
				// Log the final stats
				responseTime := time.Now()
				latencyMs := int(responseTime.Sub(startTime).Milliseconds())

				log := &model.RequestLog{
					ChannelID:     channelID,
					RequestID:     requestID,
					ModelName:     req.Model,
					UpstreamModel: upstreamModel,
					InputTokens:   inputTokens,
					OutputTokens:  outputTokens,
					TotalTokens:   inputTokens + outputTokens,
					RequestTime:   startTime,
					ResponseTime:  &responseTime,
					LatencyMs:     latencyMs,
					Status:        "success",
					IPAddress:     ipAddress,
				}
				s.logRepo.Create(log)
			}
		}

		// Extract token usage from message_start event
		if usage, ok := event["message"].(map[string]interface{}); ok {
			if u, ok := usage["usage"].(map[string]interface{}); ok {
				if input, ok := u["input_tokens"].(float64); ok {
					inputTokens = int(input)
				}
			}
		}

		// Track output tokens from content_block_delta events
		if _, ok := event["type"].(string); ok {
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if _, ok := delta["type"].(string); ok {
					outputTokens++
				}
			}
		}

		// Write to response
		jsonBytes, _ := json.Marshal(event)
		w.Write(jsonBytes)
		w.Write([]byte("\n"))
		flusher.Flush()
	}

	return nil
}

// logSuccess logs a successful request
func (s *ProxyService) logSuccess(channelID int64, requestID, modelName, upstreamModel string, startTime time.Time, resp *model.AnthropicMessageResponse, ipAddress string) {
	responseTime := time.Now()
	latencyMs := int(responseTime.Sub(startTime).Milliseconds())

	log := &model.RequestLog{
		ChannelID:     channelID,
		RequestID:     requestID,
		ModelName:     modelName,
		UpstreamModel: upstreamModel,
		InputTokens:   resp.Usage.InputTokens,
		OutputTokens:  resp.Usage.OutputTokens,
		TotalTokens:   resp.Usage.InputTokens + resp.Usage.OutputTokens,
		RequestTime:   startTime,
		ResponseTime:  &responseTime,
		LatencyMs:     latencyMs,
		Status:        "success",
		IPAddress:     ipAddress,
	}

	if _, err := s.logRepo.Create(log); err != nil {
		logger.Error("Failed to create log: %v", err)
	}
}

// logError logs a failed request
func (s *ProxyService) logError(channelID int64, requestID, modelName, upstreamModel string, startTime time.Time, httpResp *http.Response, errorCode, errorMessage, ipAddress string) {
	responseTime := time.Now()
	latencyMs := int(responseTime.Sub(startTime).Milliseconds())

	statusCode := ""
	if httpResp != nil {
		statusCode = fmt.Sprintf("%d", httpResp.StatusCode)
	}

	log := &model.RequestLog{
		ChannelID:     channelID,
		RequestID:     requestID,
		ModelName:     modelName,
		UpstreamModel: upstreamModel,
		RequestTime:   startTime,
		ResponseTime:  &responseTime,
		LatencyMs:     latencyMs,
		Status:        "error",
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
		IPAddress:     ipAddress,
	}

	if statusCode != "" {
		log.ErrorCode = statusCode + ":" + errorCode
	}

	if _, err := s.logRepo.Create(log); err != nil {
		logger.Error("Failed to create error log: %v", err)
	}
}

// TestChannel tests a channel connection
func (s *ProxyService) TestChannel(baseURL, apiKey string) error {
	req := &model.AnthropicMessageRequest{
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 10,
		Messages: []model.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	bodyBytes, _ := json.Marshal(req)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("status %d: %s", httpResp.StatusCode, string(respBody))
	}

	return nil
}
