# Local Tools

This directory contains all local tool definitions for the agent. Each tool is implemented in its own file following a consistent pattern.

## Architecture

- `interface.go` - Defines the `ToolDefinition` interface and `Registry` for managing tools
- `weather.go` - Example implementation of the weather tool
- `example_new_tool.go.template` - Template for creating new tools

## Adding a New Tool

To add a new tool, follow these steps:

### 1. Create the Tool File

Copy the template and rename it:
```bash
cp example_new_tool.go.template your_tool_name.go
```

### 2. Implement the Tool

Edit your new file and:

1. **Update the structs**: Replace `ExampleRequest` and `ExampleResponse` with your tool's input/output structures
2. **Update the tool struct**: Replace `ExampleTool` with your tool name (e.g., `CalculatorTool`)
3. **Implement the Name() method**: Return your tool's unique name
4. **Implement the Description() method**: Provide a clear description of what your tool does
5. **Implement the Define() method**: Add your tool's logic in the function

### 3. Register the Tool

Add your tool to the registry in `interface.go`:

```go
func GetDefaultRegistry() *Registry {
    registry := NewRegistry()
    
    // Register all available tools
    registry.Register(NewWeatherTool())
    registry.Register(NewYourToolTool()) // Add this line
    
    return registry
}
```

### 4. Test Your Tool

Create tests in your tool file or add them to `tool_test.go`.

## Example Tool Structure

```go
package tool

import (
    "context"
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
    "go.uber.org/zap"
)

// CalculatorRequest represents the input for calculator tool
type CalculatorRequest struct {
    Expression string `json:"expression" jsonschema:"required,description=Mathematical expression to calculate"`
}

// CalculatorResponse represents the output of calculator tool
type CalculatorResponse struct {
    Result float64 `json:"result" jsonschema:"description=The calculated result"`
}

// CalculatorTool implements a simple calculator
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
    return &CalculatorTool{}
}

func (c *CalculatorTool) Name() string {
    return "calculator"
}

func (c *CalculatorTool) Description() string {
    return "Evaluate simple mathematical expressions"
}

func (c *CalculatorTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
    return genkit.DefineTool(client, c.Name(), c.Description(),
        func(toolCtx *ai.ToolContext, input CalculatorRequest) (CalculatorResponse, error) {
            log.L(ctx).Info("calculator tool called",
                zap.String("expression", input.Expression))

            // Your calculation logic here
            result := 42.0 // Replace with actual calculation
            
            response := CalculatorResponse{
                Result: result,
            }

            log.L(ctx).Info("calculator tool response",
                zap.Float64("result", response.Result))

            return response, nil
        },
    )
}
```

## Benefits of This Architecture

1. **Modularity**: Each tool is self-contained in its own file
2. **Consistency**: All tools follow the same interface pattern
3. **Easy Registration**: Tools are automatically registered via the registry
4. **Testability**: Each tool can be tested independently
5. **Scalability**: Adding new tools requires minimal changes to existing code

## Current Tools

- `get_weather` - Returns fixed weather information ("天气晴朗")

## Future Enhancements

- Tool versioning support
- Dynamic tool loading
- Tool categories/grouping
- Tool dependency management