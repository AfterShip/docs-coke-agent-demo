package agent

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/logger"
	"github.com/firebase/genkit/go/plugins/mcp"
)

func (service *Agent) initMCP(ctx context.Context) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		logger.FromContext(ctx).Error("Config is not initialized")
		return
	}

	managerOptions := cfg.MCP.ToMCPManagerOptions()
	manager, _ := mcp.NewMCPManager(managerOptions)

	// Get tools and generate response
	tools, _ := manager.GetActiveTools(ctx, service.GenkitClient)
	for _, tool := range tools {
		logger.FromContext(ctx).Debug("Found MCP tools", "name", tool.Name())
	}

	var toolRefs []ai.ToolRef
	for _, tool := range tools {
		toolRefs = append(toolRefs, tool)
	}
	service.activeTools = toolRefs
}
