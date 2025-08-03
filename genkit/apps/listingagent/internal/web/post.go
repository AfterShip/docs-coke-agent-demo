package web

import (
	"context"
	agentmodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/agent"
	basemodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/code"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/sse"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/firebase/genkit/go/core"
	"github.com/gin-gonic/gin"
	"github.com/mingyuans/errors"
	"go.uber.org/zap"
)

func (ctrl *AgentController) PostMessages(c *gin.Context) {
	ctx := c.Request.Context()

	var req agentmodel.ProductListingInput
	if err := c.ShouldBindJSON(&req); err != nil {
		log.L(ctx).Error("bind messages request failed", zap.Error(err))
		server.NewRestfulResponseBuilder(c).
			Error(errors.WithCode(basemodel.ErrBind, "bind messages request failed: %v", err)).
			SendJSON()
		return
	}

	log.L(ctx).Debug("received messages request",
		zap.Int("message_count", len(req.Messages)))

	stream := req.Stream || c.GetHeader("Accept") == "text/event-stream"

	var msgStream *sse.MessageStream
	var callback core.StreamCallback[string]
	if stream {
		msgStream = sse.NewMessageStream(c)
		callback = func(ctx context.Context, content string) error {
			return msgStream.WriteChunk(content)
		}
	} else {
		c.Header("Content-Type", "application/json")
	}

	content, err := ctrl.Agent.SimpleRun(ctx, req, callback)
	if err != nil {
		log.L(ctx).Error("failed to run agent", zap.Error(err))
		server.NewRestfulResponseBuilder(c).
			Error(errors.WithCode(basemodel.ErrUnknown, "failed to run agent: %v", err)).
			SendJSON()
		return
	}

	if !stream {
		// If not streaming, send the final content as JSON response
		server.NewRestfulResponseBuilder(c).
			Data(content).
			SendJSON()
		return
	}

	// Finish the stream
	if err := msgStream.Finish(); err != nil {
		log.L(ctx).Error("failed to finish message stream", zap.Error(err))
		return
	}

	log.L(ctx).Info("message stream completed",
		zap.String("message_id", msgStream.GetMessageID()))
}
