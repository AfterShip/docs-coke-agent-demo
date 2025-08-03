package web

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/agent"
)

// AgentController handles requests related to AI agents.
type AgentController struct {
	Agent *agent.Agent
}

// NewAgentController creates a new instance of AgentController.
func NewAgentController(cfg config.Config) *AgentController {
	return &AgentController{
		Agent: agent.NewAgent(cfg),
	}
}
