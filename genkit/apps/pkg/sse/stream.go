package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/uuid"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
)

// EventType defines SSE event types
type EventType string

const (
	EventMessageStart      EventType = "message_start"
	EventContentBlockStart EventType = "content_block_start"
	EventPing              EventType = "ping"
	EventContentBlockDelta EventType = "content_block_delta"
	EventContentBlockStop  EventType = "content_block_stop"
	EventMessageDelta      EventType = "message_delta"
	EventMessageStop       EventType = "message_stop"
	EventError             EventType = "error"
)

// StreamWriter handles SSE streaming
type StreamWriter struct {
	ctx    context.Context
	writer gin.ResponseWriter
}

// NewStreamWriter creates a new SSE stream writer
func NewStreamWriter(c *gin.Context) *StreamWriter {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	c.Header("Transfer-Encoding", "chunked")

	return &StreamWriter{
		ctx:    c.Request.Context(),
		writer: c.Writer,
	}
}

// WriteEvent sends an SSE event
func (sw *StreamWriter) WriteEvent(eventType EventType, data interface{}) error {
	// Check if client disconnected
	select {
	case <-sw.ctx.Done():
		log.L(sw.ctx).Info("client disconnected during event write")
		return sw.ctx.Err()
	default:
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.L(sw.ctx).Error("failed to marshal SSE data", zap.Error(err))
		return err
	}

	_, err = fmt.Fprintf(sw.writer, "event: %s\ndata: %s\n\n", string(eventType), string(jsonData))
	if err != nil {
		log.L(sw.ctx).Error("failed to write SSE data", zap.Error(err))
		return err
	}

	sw.writer.Flush()
	return nil
}

// MessageStream manages the complete SSE message streaming flow
type MessageStream struct {
	writer    *StreamWriter
	messageID string
	once      sync.Once
}

// NewMessageStream creates a new message stream
func NewMessageStream(c *gin.Context) *MessageStream {
	messageID := uuid.GenerateUUIDV4()
	return &MessageStream{
		writer:    NewStreamWriter(c),
		messageID: messageID,
		once:      sync.Once{},
	}
}

// Start sends the initial message_start, content_block_start, and ping events
func (ms *MessageStream) Start() error {
	// Send message_start event
	messageStart := MessageStartData{
		Type: string(EventMessageStart),
		Message: ResponseMessage{
			ID:      ms.messageID,
			Type:    "message",
			Role:    "model",
			Content: []string{},
		},
	}
	if err := ms.writer.WriteEvent(EventMessageStart, messageStart); err != nil {
		return err
	}

	// Send content_block_start event
	contentBlockStart := ContentBlockStartData{
		Type:  string(EventContentBlockStart),
		Index: 0,
		ContentBlock: ContentBlock{
			Type: "text",
			Text: "",
		},
	}
	if err := ms.writer.WriteEvent(EventContentBlockStart, contentBlockStart); err != nil {
		return err
	}

	// Send ping event
	ping := PingData{
		Type: string(EventPing),
	}
	return ms.writer.WriteEvent(EventPing, ping)
}

// WriteChunk sends a content_block_delta event
func (ms *MessageStream) WriteChunk(chunk string) error {
	ms.once.Do(func() {
		_ = ms.Start()
	})

	contentDelta := ContentBlockDeltaData{
		Type:  string(EventContentBlockDelta),
		Index: 0,
		Delta: Delta{
			Type: "text_delta",
			Text: chunk,
		},
	}
	return ms.writer.WriteEvent(EventContentBlockDelta, contentDelta)
}

// Finish sends the final events to complete the stream
func (ms *MessageStream) Finish() error {
	// Send content_block_stop event
	contentBlockStop := ContentBlockStopData{
		Type:  string(EventContentBlockStop),
		Index: 0,
	}
	if err := ms.writer.WriteEvent(EventContentBlockStop, contentBlockStop); err != nil {
		return err
	}

	// Send message_delta event
	stopReason := "end_turn"
	messageDelta := MessageDeltaData{
		Type: string(EventMessageDelta),
		Delta: Delta{
			Type:         "message_delta",
			StopReason:   &stopReason,
			StopSequence: nil,
		},
	}
	if err := ms.writer.WriteEvent(EventMessageDelta, messageDelta); err != nil {
		return err
	}

	// Send message_stop event
	messageStop := MessageStopData{
		Type: string(EventMessageStop),
	}
	return ms.writer.WriteEvent(EventMessageStop, messageStop)
}

// GetMessageID returns the generated message ID
func (ms *MessageStream) GetMessageID() string {
	return ms.messageID
}

// StreamCallback returns a callback function for streaming
func (ms *MessageStream) StreamCallback() func(string, bool) error {
	return func(chunk string, isComplete bool) error {
		return ms.WriteChunk(chunk)
	}
}
