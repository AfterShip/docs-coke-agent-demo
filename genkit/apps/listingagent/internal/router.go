package internal

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	controller_agent "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/web"
	"github.com/gin-gonic/gin"
)

// InstallControllers install APIs
func InstallControllers(g *gin.Engine, cfg config.Config) {
	v1Group := g.Group("/v1")
	installAgentController(v1Group, cfg)
}

func installAgentController(g *gin.RouterGroup, cfg config.Config) {
	v1Agent := g.Group("/agent")
	ctl := controller_agent.NewAgentController(cfg)
	//另外一种写法
	//flows := genkit.ListFlows(ctl.Agent.GetGenkitClient())
	//for _, flow := range flows {
	//	v1Agent.POST("/"+flow.Name(), func(c *gin.Context) {
	//		genkit.Handler(flow)(c.Writer, c.Request)
	//	})
	//}
	v1Agent.POST("/messages", ctl.PostMessages)
}
