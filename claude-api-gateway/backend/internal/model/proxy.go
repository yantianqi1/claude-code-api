package model

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type string `json:"type"`
	Delta string `json:"delta,omitempty"`
}

// AnthropicMessageRequest represents the Anthropic Messages API request
type AnthropicMessageRequest struct {
	Model            string    `json:"model"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []Message `json:"messages"`
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	TopK             *int      `json:"top_k,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	StopSequences    []string  `json:"stop_sequences,omitempty"`
	System           string    `json:"system,omitempty"`
	Tools            []Tool    `json:"tools,omitempty"`
	ToolChoice       any       `json:"tool_choice,omitempty"`
	Metadata         any       `json:"metadata,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AnthropicMessageResponse represents the Anthropic Messages API response
type AnthropicMessageResponse struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Role         string   `json:"role"`
	Content      []Content `json:"content"`
	Model        string   `json:"model"`
	StopReason   string   `json:"stop_reason"`
	Usage        Usage    `json:"usage"`
}

// Content represents content in the response
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicErrorResponse represents an error response from Anthropic API
type AnthropicErrorResponse struct {
	Type  string `json:"type"`
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	Type         string `json:"type"`
	Index        int    `json:"index,omitempty"`
	Delta        *Delta `json:"delta,omitempty"`
	Message      *AnthropicMessageResponse `json:"message,omitempty"`
	Usage        *Usage `json:"usage,omitempty"`
}

// Delta represents text delta in streaming
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
