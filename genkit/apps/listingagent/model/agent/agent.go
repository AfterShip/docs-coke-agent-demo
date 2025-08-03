package agentmodel

import (
	"github.com/AfterShip/connectors-library/sdks/product_listings"
	"github.com/firebase/genkit/go/ai"
)

// Message represents a single message in the conversation.
type Message struct {
	// Role of the message sender (user or assistant).
	Role ai.Role `json:"role"`
	// Content of the message.
	Content string `json:"content"`
}

type ProductListingInput struct {
	Messages []Message       `json:"messages"`
	Metadata ListingMetadata `json:"metadata"`
	// Stream indicates whether to stream the response.
	Stream bool `json:"stream"`
}

type ListingMetadata struct {
	Command string     `json:"command,omitempty"`
	Plan    Plan       `json:"plan,omitempty"`
	Context JobContext `json:"context,omitempty"` // New job context field
}

type ProductListingOutput struct {
	Message  Message         `json:"message"`
	Metadata ListingMetadata `json:"metadata"`
}

type Plan struct {
	ID         string          `json:"id,omitempty"`
	Command    string          `json:"command,omitempty"` // publish|query|edit|activate|deactivate
	Edit       []EditArg       `json:"edits,omitempty"`
	Publish    []PublishArg    `json:"publishes,omitempty"`
	Query      []QueryArg      `json:"queries,omitempty"`
	Activate   []ActivateArg   `json:"activates,omitempty"`
	Deactivate []DeactivateArg `json:"deactivates,omitempty"`
}

type EditArg struct {
	ProductListingID string                   `json:"product_listing_id"`
	Product          product_listings.Product `json:"product"`
}

type PublishArg struct {
	ProductCenterID string `json:"product_center_id"`
}

type QueryArg struct {
	ProductListingID string `json:"product_listing_id"`
}

type ActivateArg struct {
	ProductListingID string `json:"product_listing_id"`
}

type DeactivateArg struct {
	ProductListingID string `json:"product_listing_id"`
}

// JobContext contains the Job object for metadata passing
type JobContext struct {
	Job *Job `json:"job,omitempty"` // Job object stored in metadata
}

// Job represents a task with progressive information collection
type Job struct {
	ID      string                 `json:"id"`              // Unique job identifier
	Type    string                 `json:"type"`            // Fixed as "product_listing"
	Phase   Phase                  `json:"phase"`           // Current phase: reasoning|acting|completed|failed
	Intent  string                 `json:"intent"`          // User intent: publish|query|edit|activate|deactivate
	Context map[string]interface{} `json:"context"`         // Collected information and state
	Error   string                 `json:"error,omitempty"` // Error message if any
}

// Phase represents the current processing phase
type Phase string

const (
	PhaseReasoning Phase = "reasoning" // Reasoning phase: analyze intent, collect information
	PhaseActing    Phase = "acting"    // Acting phase: execute specific operations
	PhaseCompleted Phase = "completed" // Completed
	PhaseFailed    Phase = "failed"    // Failed
)
