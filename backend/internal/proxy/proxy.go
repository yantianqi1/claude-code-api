package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	// Track if we've started writing to response
	// Once headers are written, we cannot try another channel
	var lastErr error

	// Try each mapped channel
	for _, mapping := range mappings {
		if !mapping.IsEnabled {
			continue
		}

		channel, err := s.channelRepo.GetByID(mapping.ChannelID)
		if err != nil || !channel.IsActive {
			continue
		}

		err = s.proxyStreamToChannel(req, apiKey, ipAddress, requestID, startTime, channel, mapping.UpstreamModel, w)
		if err != nil {
			logger.Error("Stream proxy to channel %d failed: %v", channel.ID, err)
			lastErr = err
			// Check if response has been started (headers written)
			// If so, we cannot try another channel
			if rw, ok := w.(interface{ Written() bool }); ok && rw.Written() {
				// Response already started, cannot switch channels
				return err
			}
			continue
		}
		return nil
	}

	if lastErr != nil {
		return lastErr
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

	bodyBytes, err := json.Marshal(proxyReq)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKeyToUse)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "http_request", err.Error(), ipAddress)
		return err
	}
	defer httpResp.Body.Close()

	// Check for error response before streaming
	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		var errResp model.AnthropicErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			s.logError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, errResp.Error.Type, errResp.Error.Message, ipAddress)
			return fmt.Errorf("API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		s.logError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "unknown", string(respBody), ipAddress)
		return fmt.Errorf("API error: status %d, response: %s", httpResp.StatusCode, string(respBody))
	}

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

	// Use SSE line scanner instead of JSON decoder
	// Anthropic streaming uses SSE format: "event: xxx\ndata: {...}\n\n"
	scanner := newLineScanner(httpResp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle event type line (optional)
		if strings.HasPrefix(line, "event:") {
			// Forward the event line as-is
			w.Write([]byte(line))
			w.Write([]byte("\n"))
			continue
		}

		// Handle data line
		if strings.HasPrefix(line, "data:") {
			// Extract JSON data
			dataStr := strings.TrimPrefix(line, "data:")
			dataStr = strings.TrimSpace(dataStr)

			if dataStr == "" {
				continue
			}

			// Parse JSON to track token usage
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &event); err == nil {
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
				if msg, ok := event["message"].(map[string]interface{}); ok {
					if u, ok := msg["usage"].(map[string]interface{}); ok {
						if input, ok := u["input_tokens"].(float64); ok {
							inputTokens = int(input)
						}
					}
				}

				// Extract output tokens from message_delta event
				if eventType, ok := event["type"].(string); ok {
					if eventType == "message_delta" {
						if usage, ok := event["usage"].(map[string]interface{}); ok {
							if output, ok := usage["output_tokens"].(float64); ok {
								outputTokens = int(output)
							}
						}
					}
				}
			}

			// Forward the data line
			w.Write([]byte(line))
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
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

// ProxyChat proxies an OpenAI format chat completion request
func (s *ProxyService) ProxyChat(req *model.OpenAIChatRequest, apiKey string, ipAddress string) (*model.OpenAIChatResponse, error) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Find model mapping
	mappings, err := s.mappingRepo.FindByDisplayModel(req.Model)
	if err != nil || len(mappings) == 0 {
		// No mapping found, proxy directly
		return s.proxyChatToChannel(req, apiKey, ipAddress, requestID, startTime, nil, req.Model)
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

		resp, err := s.proxyChatToChannel(req, apiKey, ipAddress, requestID, startTime, channel, mapping.UpstreamModel)
		if err != nil {
			logger.Error("Failed to proxy chat to channel %d: %v", channel.ID, err)
			continue
		}
		return resp, nil
	}

	return nil, fmt.Errorf("all channels failed for model: %s", req.Model)
}

// proxyChatToChannel proxies OpenAI chat request to a specific channel
func (s *ProxyService) proxyChatToChannel(req *model.OpenAIChatRequest, apiKey string, ipAddress string, requestID string, startTime time.Time, channel *model.Channel, upstreamModel string) (*model.OpenAIChatResponse, error) {
	var channelID int64
	var baseURL string
	var apiKeyToUse string
	var timeout time.Duration
	var provider string

	if channel != nil {
		channelID = channel.ID
		baseURL = channel.BaseURL
		apiKeyToUse = channel.APIKey
		timeout = time.Duration(channel.Timeout) * time.Second
		provider = channel.Provider
	} else {
		channelID = 0
		baseURL = "https://api.anthropic.com"
		apiKeyToUse = apiKey
		timeout = 120 * time.Second
		provider = "anthropic"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var httpReq *http.Request
	var err error

	// Choose API format based on provider
	if provider == "anthropic" {
		// Convert OpenAI format to Anthropic format
		anthropicReq := s.convertOpenAIToAnthropic(req, upstreamModel)

		bodyBytes, err := json.Marshal(anthropicReq)
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err = http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", apiKeyToUse)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	} else {
		// Use OpenAI format directly
		proxyReq := &model.OpenAIChatRequest{
			Model:       upstreamModel,
			Messages:    req.Messages,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			Stream:      false,
			Stop:        req.Stop,
			Tools:       req.Tools,
			ToolChoice:  req.ToolChoice,
		}

		if req.MaxTokens != nil {
			proxyReq.MaxTokens = req.MaxTokens
		} else {
			defaultTokens := 4096
			proxyReq.MaxTokens = &defaultTokens
		}

		bodyBytes, err := json.Marshal(proxyReq)
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err = http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKeyToUse)
	}

	// Make request
	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "http_request", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "read_response", err.Error(), ipAddress)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if httpResp.StatusCode != http.StatusOK {
		s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "api_error", string(respBody), ipAddress)
		return nil, fmt.Errorf("API error: status %d, response: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response based on provider
	var response *model.OpenAIChatResponse
	if provider == "anthropic" {
		// Convert Anthropic response to OpenAI format
		var anthropicResp model.AnthropicMessageResponse
		if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "parse_response", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		response = s.convertAnthropicToOpenAI(&anthropicResp, req.Model)
	} else {
		if err := json.Unmarshal(respBody, &response); err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "parse_response", err.Error(), ipAddress)
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	// Log successful request
	s.logChatSuccess(channelID, requestID, req.Model, upstreamModel, startTime, response, ipAddress)

	return response, nil
}

// convertOpenAIToAnthropic converts OpenAI chat request to Anthropic format
func (s *ProxyService) convertOpenAIToAnthropic(req *model.OpenAIChatRequest, upstreamModel string) *model.AnthropicMessageRequest {
	var messages []model.Message
	var systemPrompt string

	for _, msg := range req.Messages {
		role := msg.Role
		content := ""

		// Extract content as string
		switch c := msg.Content.(type) {
		case string:
			content = c
		case []interface{}:
			// Handle array of content blocks
			for _, block := range c {
				if m, ok := block.(map[string]interface{}); ok {
					if t, ok := m["text"].(string); ok {
						content += t
					}
				}
			}
		}

		// Handle system message
		if role == "system" {
			systemPrompt = content
			continue
		}

		// Map OpenAI roles to Anthropic roles
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}

		messages = append(messages, model.Message{
			Role:    role,
			Content: content,
		})
	}

	maxTokens := 4096
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	return &model.AnthropicMessageRequest{
		Model:       upstreamModel,
		MaxTokens:   maxTokens,
		Messages:    messages,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		System:      systemPrompt,
	}
}

// convertAnthropicToOpenAI converts Anthropic response to OpenAI format
func (s *ProxyService) convertAnthropicToOpenAI(resp *model.AnthropicMessageResponse, displayModel string) *model.OpenAIChatResponse {
	content := ""
	for _, c := range resp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	finishReason := "stop"
	if resp.StopReason == "max_tokens" {
		finishReason = "length"
	}

	return &model.OpenAIChatResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   displayModel,
		Choices: []model.OpenAIChoice{
			{
				Index: 0,
				Message: model.OpenAIMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		},
		Usage: model.OpenAIUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// ProxyChatStream proxies an OpenAI streaming chat completion request
func (s *ProxyService) ProxyChatStream(req *model.OpenAIChatRequest, apiKey string, ipAddress string, w http.ResponseWriter) error {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Find model mapping
	mappings, err := s.mappingRepo.FindByDisplayModel(req.Model)
	if err != nil || len(mappings) == 0 {
		return s.proxyChatStreamToChannel(req, apiKey, ipAddress, requestID, startTime, nil, req.Model, w)
	}

	// Track if we've started writing to response
	var lastErr error

	// Try each mapped channel
	for _, mapping := range mappings {
		if !mapping.IsEnabled {
			continue
		}

		channel, err := s.channelRepo.GetByID(mapping.ChannelID)
		if err != nil || !channel.IsActive {
			continue
		}

		err = s.proxyChatStreamToChannel(req, apiKey, ipAddress, requestID, startTime, channel, mapping.UpstreamModel, w)
		if err != nil {
			logger.Error("Chat stream proxy to channel %d failed: %v", channel.ID, err)
			lastErr = err
			// Check if response has been started (headers written)
			if rw, ok := w.(interface{ Written() bool }); ok && rw.Written() {
				return err
			}
			continue
		}
		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("all channels failed for streaming model: %s", req.Model)
}

// proxyChatStreamToChannel proxies OpenAI streaming request to a specific channel
func (s *ProxyService) proxyChatStreamToChannel(req *model.OpenAIChatRequest, apiKey string, ipAddress string, requestID string, startTime time.Time, channel *model.Channel, upstreamModel string, w http.ResponseWriter) error {
	var channelID int64
	var baseURL string
	var apiKeyToUse string
	var timeout time.Duration
	var provider string

	if channel != nil {
		channelID = channel.ID
		baseURL = channel.BaseURL
		apiKeyToUse = channel.APIKey
		timeout = time.Duration(channel.Timeout) * time.Second
		provider = channel.Provider
	} else {
		channelID = 0
		baseURL = "https://api.anthropic.com"
		apiKeyToUse = apiKey
		timeout = 120 * time.Second
		provider = "anthropic"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var httpReq *http.Request
	var err error

	// Choose API format based on provider
	if provider == "anthropic" {
		// Convert OpenAI format to Anthropic format
		anthropicReq := s.convertOpenAIToAnthropic(req, upstreamModel)
		anthropicReq.Stream = true

		bodyBytes, err := json.Marshal(anthropicReq)
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err = http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/messages", bytes.NewReader(bodyBytes))
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
			return fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("x-api-key", apiKeyToUse)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
	} else {
		// Use OpenAI format directly
		proxyReq := &model.OpenAIChatRequest{
			Model:       upstreamModel,
			Messages:    req.Messages,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			Stream:      true,
			Stop:        req.Stop,
			Tools:       req.Tools,
			ToolChoice:  req.ToolChoice,
		}

		if req.MaxTokens != nil {
			proxyReq.MaxTokens = req.MaxTokens
		} else {
			defaultTokens := 4096
			proxyReq.MaxTokens = &defaultTokens
		}

		bodyBytes, err := json.Marshal(proxyReq)
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "marshal_request", err.Error(), ipAddress)
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err = http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		if err != nil {
			s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "create_request", err.Error(), ipAddress)
			return fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKeyToUse)
	}

	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, nil, "http_request", err.Error(), ipAddress)
		return err
	}
	defer httpResp.Body.Close()

	// Check for error response before streaming
	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		s.logChatError(channelID, requestID, req.Model, upstreamModel, startTime, httpResp, "api_error", string(respBody), ipAddress)
		return fmt.Errorf("API error: status %d, response: %s", httpResp.StatusCode, string(respBody))
	}

	// Set streaming headers after upstream validation
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Copy stream
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	// Handle streaming based on provider
	hasData := false
	if provider == "anthropic" {
		// Convert Anthropic SSE stream to OpenAI SSE format
		hasData = s.convertAnthropicStreamToOpenAI(httpResp.Body, w, flusher, req.Model, requestID)
	} else {
		// Forward OpenAI SSE stream directly
		scanner := newLineScanner(httpResp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if strings.HasPrefix(line, ":") {
				continue
			}

			if strings.HasPrefix(line, "data:") {
				hasData = true
				w.Write([]byte(line))
				w.Write([]byte("\n\n"))
				flusher.Flush()
			}
		}
	}

	// Log success for streaming
	if hasData {
		responseTime := time.Now()
		latencyMs := int(responseTime.Sub(startTime).Milliseconds())
		log := &model.RequestLog{
			ChannelID:     channelID,
			RequestID:     requestID,
			ModelName:     req.Model,
			UpstreamModel: upstreamModel,
			RequestTime:   startTime,
			ResponseTime:  &responseTime,
			LatencyMs:     latencyMs,
			Status:        "success",
			IPAddress:     ipAddress,
		}
		s.logRepo.Create(log)
	}

	return nil
}

// convertAnthropicStreamToOpenAI converts Anthropic SSE stream to OpenAI SSE format
func (s *ProxyService) convertAnthropicStreamToOpenAI(body io.Reader, w http.ResponseWriter, flusher http.Flusher, displayModel string, requestID string) bool {
	scanner := newLineScanner(body)
	hasData := false
	messageID := "chatcmpl-" + requestID

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimPrefix(line, "data:")
			dataStr = strings.TrimSpace(dataStr)

			if dataStr == "" {
				continue
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
				continue
			}

			eventType, _ := event["type"].(string)

			switch eventType {
			case "content_block_delta":
				// Extract text delta
				if delta, ok := event["delta"].(map[string]interface{}); ok {
					if text, ok := delta["text"].(string); ok {
						hasData = true
						chunk := model.OpenAIStreamChunk{
							ID:      messageID,
							Object:  "chat.completion.chunk",
							Created: time.Now().Unix(),
							Model:   displayModel,
							Choices: []model.OpenAIStreamChoice{
								{
									Index: 0,
									Delta: model.OpenAIDelta{
										Content: text,
									},
									FinishReason: nil,
								},
							},
						}
						chunkBytes, _ := json.Marshal(chunk)
						w.Write([]byte("data: "))
						w.Write(chunkBytes)
						w.Write([]byte("\n\n"))
						flusher.Flush()
					}
				}

			case "message_start":
				// Send initial chunk with role
				hasData = true
				chunk := model.OpenAIStreamChunk{
					ID:      messageID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   displayModel,
					Choices: []model.OpenAIStreamChoice{
						{
							Index: 0,
							Delta: model.OpenAIDelta{
								Role: "assistant",
							},
							FinishReason: nil,
						},
					},
				}
				chunkBytes, _ := json.Marshal(chunk)
				w.Write([]byte("data: "))
				w.Write(chunkBytes)
				w.Write([]byte("\n\n"))
				flusher.Flush()

			case "message_stop":
				// Send final chunk with finish_reason
				finishReason := "stop"
				chunk := model.OpenAIStreamChunk{
					ID:      messageID,
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   displayModel,
					Choices: []model.OpenAIStreamChoice{
						{
							Index:        0,
							Delta:        model.OpenAIDelta{},
							FinishReason: &finishReason,
						},
					},
				}
				chunkBytes, _ := json.Marshal(chunk)
				w.Write([]byte("data: "))
				w.Write(chunkBytes)
				w.Write([]byte("\n\n"))
				w.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
			}
		}
	}

	return hasData
}

// logChatSuccess logs a successful OpenAI chat request
func (s *ProxyService) logChatSuccess(channelID int64, requestID, modelName, upstreamModel string, startTime time.Time, resp *model.OpenAIChatResponse, ipAddress string) {
	responseTime := time.Now()
	latencyMs := int(responseTime.Sub(startTime).Milliseconds())

	log := &model.RequestLog{
		ChannelID:     channelID,
		RequestID:     requestID,
		ModelName:     modelName,
		UpstreamModel: upstreamModel,
		InputTokens:   resp.Usage.PromptTokens,
		OutputTokens:  resp.Usage.CompletionTokens,
		TotalTokens:   resp.Usage.TotalTokens,
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

// logChatError logs a failed OpenAI chat request
func (s *ProxyService) logChatError(channelID int64, requestID, modelName, upstreamModel string, startTime time.Time, httpResp *http.Response, errorCode, errorMessage, ipAddress string) {
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

// lineScanner helps scan SSE streams line by line
type lineScanner struct {
	reader io.Reader
	buffer []byte
}

func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{reader: r, buffer: make([]byte, 0, 4096)}
}

func (s *lineScanner) Scan() bool {
	buf := make([]byte, 1)
	s.buffer = s.buffer[:0]

	for {
		n, err := s.reader.Read(buf)
		if err != nil {
			return len(s.buffer) > 0
		}
		if n > 0 {
			if buf[0] == '\n' {
				return len(s.buffer) > 0
			}
			if buf[0] != '\r' {
				s.buffer = append(s.buffer, buf[0])
			}
		}
	}
}

func (s *lineScanner) Text() string {
	return string(s.buffer)
}

func (s *lineScanner) Bytes() []byte {
	return s.buffer
}
