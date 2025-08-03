package agent

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/agent/tools"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"go.uber.org/zap"
)

// initLocalTools initializes local tools using the tool registry and adds them to activeTools
func (service *Agent) initLocalTools(ctx context.Context) {
	// Get the default tool registry with all available tools
	registry := tools.GetDefaultRegistry()

	// Define all tools from the registry
	localTools := registry.DefineAll(ctx, service.GenkitClient)

	// Add local tools to existing activeTools
	service.activeTools = append(service.activeTools, localTools...)

	// Log tool registration details
	registeredTools := registry.GetRegisteredTools()
	toolNames := make([]string, len(registeredTools))
	for i, t := range registeredTools {
		toolNames[i] = t.Name()
	}

	log.L(ctx).Info("Local tools initialized",
		zap.Strings("tools", toolNames),
		zap.Int("count", len(localTools)),
		zap.Int("totalActiveTools", len(service.activeTools)))
}
