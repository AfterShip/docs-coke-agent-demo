package sse

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ResponseMessage represents the message structure in SSE responses
type ResponseMessage struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Role         string   `json:"role"`
	Content      []string `json:"content"`
	Model        string   `json:"model"`
	StopReason   *string  `json:"stop_reason"`
	StopSequence *string  `json:"stop_sequence"`
	Usage        Usage    `json:"usage"`
}

// ContentBlock represents a content block in SSE events
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Delta represents delta information in SSE events
type Delta struct {
	Type         string  `json:"type"`
	Text         string  `json:"text,omitempty"`
	StopReason   *string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

// MessageStartData represents the message_start event data
type MessageStartData struct {
	Type    string          `json:"type"`
	Message ResponseMessage `json:"message"`
}

// ContentBlockStartData represents the content_block_start event data
type ContentBlockStartData struct {
	Type         string       `json:"type"`
	Index        int          `json:"index"`
	ContentBlock ContentBlock `json:"content_block"`
}

// PingData represents the ping event data
type PingData struct {
	Type string `json:"type"`
}

// ContentBlockDeltaData represents the content_block_delta event data
type ContentBlockDeltaData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

// ContentBlockStopData represents the content_block_stop event data
type ContentBlockStopData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// MessageDeltaData represents the message_delta event data
type MessageDeltaData struct {
	Type  string `json:"type"`
	Delta Delta  `json:"delta"`
	Usage Usage  `json:"usage"`
}

// MessageStopData represents the message_stop event data
type MessageStopData struct {
	Type string `json:"type"`
}

// ErrorData represents the error event data
type ErrorData struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
