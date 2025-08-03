package tools

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// ToolDefinition represents a local tool definition
type ToolDefinition interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Define creates and registers the tool with the given genkit client
	Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef
}

// Registry manages all local tools
type Registry struct {
	tools []ToolDefinition
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make([]ToolDefinition, 0),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool ToolDefinition) {
	r.tools = append(r.tools, tool)
}

// DefineAll defines all registered tools with the given genkit client
func (r *Registry) DefineAll(ctx context.Context, client *genkit.Genkit) []ai.ToolRef {
	var toolRefs []ai.ToolRef

	for _, tool := range r.tools {
		toolRef := tool.Define(ctx, client)
		toolRefs = append(toolRefs, toolRef)
	}

	return toolRefs
}

// GetRegisteredTools returns all registered tool definitions
func (r *Registry) GetRegisteredTools() []ToolDefinition {
	return r.tools
}

// GetDefaultRegistry returns a registry with all available local tools
func GetDefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register all available tools
	registry.Register(NewGetProductListingsTool())
	registry.Register(NewGetProductListingByIDTool())
	registry.Register(NewUpdateProductListingDescriptionTool())
	registry.Register(NewUpdateProductListingTitleTool())

	return registry
}
