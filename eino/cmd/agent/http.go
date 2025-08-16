package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type ProcessFunc func(ctx context.Context, request *Request) (*schema.StreamReader[*schema.Message], error)

// Anthropic streaming protocol event structures
type StreamEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type MessageStart struct {
	Type    string   `json:"type"`
	Message *Message `json:"message"`
}

type Message struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   *string        `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        *Usage         `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ContentBlockStart struct {
	Type         string        `json:"type"`
	Index        int           `json:"index"`
	ContentBlock *ContentBlock `json:"content_block"`
}

type ContentBlockDelta struct {
	Type  string     `json:"type"`
	Index int        `json:"index"`
	Delta *TextDelta `json:"delta"`
}

type TextDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ContentBlockStop struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type MessageDelta struct {
	Type  string       `json:"type"`
	Delta *MessageInfo `json:"delta"`
	Usage *Usage       `json:"usage,omitempty"`
}

type MessageInfo struct {
	StopReason   *string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

type MessageStop struct {
	Type string `json:"type"`
}

func createGinServer(processor ProcessFunc) {
	r := gin.Default()
	r.POST("/agents/hello/v1/messages", newHandler(processor))

	log.Printf("Server starting on port 8000...\n")
	log.Printf("POST endpoint available at: http://localhost:8000/agents/hello/v1/messages\n")

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func newHandler(processor ProcessFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if processor == nil {
			c.JSON(500, gin.H{"error": "Processor function is not set"})
			return
		}
		// Print request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			c.JSON(500, gin.H{"error": "Failed to read request body"})
			return
		}
		log.Printf("Received request body: %s", string(bodyBytes))
		
		// Restore body for JSON binding
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		
		var request Request
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("failed to bind request: %v", err)
			c.JSON(400, gin.H{"error": "Invalid request format"})
			return
		}
		ctx := c.Request.Context()
		streamReader, err := processor(ctx, &request)
		if err != nil {
			log.Printf("Error processing request: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		// Set SSE headers
		setupSSEResponseHeaders(c)

		// Send message_start event
		sendMessageStartEvent(c)

		// Send content_block_start event
		sendContentBlockStartEvent(c)

		// Process streaming messages
		outputTokens := 0
		for {
			msg, receiveErr := streamReader.Recv()
			if receiveErr != nil {
				if receiveErr == io.EOF {
					break
				}
				log.Printf("Stream error: %v", receiveErr)
				return
			}

			if msg != nil && msg.Content != "" {
				// Send content_block_delta event
				sendContentBlockDeltaEvent(c, msg)
				outputTokens++
			}
		}

		// Send content_block_stop event
		sendContentBlockStopEvent(c)

		// Send message_delta event with final usage
		sendMessageDeltaEvent(c, outputTokens)

		// Send message_stop event
		sendMessageStopEvent(c)
	}
}

func sendMessageStopEvent(c *gin.Context) {
	messageStop := MessageStop{
		Type: "message_stop",
	}
	sendSSEEvent(c, "message_stop", messageStop)
}

func sendMessageDeltaEvent(c *gin.Context, outputTokens int) {
	messageDelta := MessageDelta{
		Type: "message_delta",
		Delta: &MessageInfo{
			StopReason: stringPtr("end_turn"),
		},
		Usage: &Usage{
			OutputTokens: outputTokens,
		},
	}
	sendSSEEvent(c, "message_delta", messageDelta)
}

func sendContentBlockStopEvent(c *gin.Context) {
	contentBlockStop := ContentBlockStop{
		Type:  "content_block_stop",
		Index: 0,
	}
	sendSSEEvent(c, "content_block_stop", contentBlockStop)
}

func sendContentBlockDeltaEvent(c *gin.Context, msg *schema.Message) {
	// Parse the content to extract text parts
	text := extractTextFromContent(msg.Content)
	if text == "" {
		return // Skip empty content
	}
	
	contentDelta := ContentBlockDelta{
		Type:  "content_block_delta",
		Index: 0,
		Delta: &TextDelta{
			Type: "text_delta",
			Text: text,
		},
	}
	sendSSEEvent(c, "content_block_delta", contentDelta)
}

func stringPtr(s string) *string {
	return &s
}

// extractTextFromContent extracts text content from the message content
// It handles both plain text and JSON formatted content
func extractTextFromContent(content string) string {
	if content == "" {
		return ""
	}
	
	// Try to parse as JSON array of content blocks
	var contentBlocks []ContentBlock
	if err := json.Unmarshal([]byte(content), &contentBlocks); err == nil {
		// Successfully parsed as JSON, extract text from text blocks
		var textParts []string
		for _, block := range contentBlocks {
			if block.Type == "text" && block.Text != "" {
				textParts = append(textParts, block.Text)
			}
		}
		if len(textParts) > 0 {
			return strings.Join(textParts, "")
		}
	}
	
	// If not JSON or no text blocks found, treat as plain text
	return content
}

func setupSSEResponseHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusOK)
}

func sendMessageStartEvent(c *gin.Context) {
	// Send message_start event
	messageID := fmt.Sprintf("msg_%d", c.GetInt64("timestamp"))
	messageStart := MessageStart{
		Type: "message_start",
		Message: &Message{
			ID:      messageID,
			Type:    "message",
			Role:    "assistant",
			Content: []ContentBlock{},
			Model:   "claude-3-sonnet",
			Usage:   &Usage{InputTokens: 0, OutputTokens: 0},
		},
	}
	sendSSEEvent(c, "message_start", messageStart)
}

func sendContentBlockStartEvent(c *gin.Context) {
	contentBlockStart := ContentBlockStart{
		Type:  "content_block_start",
		Index: 0,
		ContentBlock: &ContentBlock{
			Type: "text",
			Text: "",
		},
	}
	sendSSEEvent(c, "content_block_start", contentBlockStart)
}

func sendSSEEvent(c *gin.Context, eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling SSE data: %v", err)
		return
	}

	c.SSEvent(eventType, string(jsonData))
	c.Writer.Flush()
}
