package agent

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/config"
	agentmodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/agent"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/code"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/llm"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/tools/prompt"
	aftership_genkit "github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/aftership-genkit"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/mingyuans/errors"
	"github.com/mingyuans/genkit-anthropic/anthropic"
	"go.uber.org/zap"
)

const PromptDirectory = "./apps/listingagent/internal/agent/prompts"

type Agent struct {
	GenkitClient *genkit.Genkit
	activeTools  []ai.ToolRef
	promptLoader *prompt.Loader
}

func NewAgent(cfg config.Config) *Agent {
	promptLoader := prompt.NewPromptLoader(cfg.LLM.PromptDirectory, nil)
	client := initGenkitClient(context.Background(), cfg.LLM)
	instance := &Agent{
		GenkitClient: client,
		activeTools:  make([]ai.ToolRef, 0),
		promptLoader: promptLoader,
	}
	ctx := context.Background()
	instance.initMCP(ctx)
	instance.initLocalTools(ctx)
	instance.initAgentFlows()
	return instance
}

func (service *Agent) GetGenkitClient() *genkit.Genkit {
	return service.GenkitClient
}

func initGenkitClient(ctx context.Context, opt *llm.Option) *genkit.Genkit {
	llm.SetupGenkitLogger()
	client, err := genkit.Init(
		ctx,
		genkit.WithPromptDir(opt.PromptDirectory),
		genkit.WithPlugins(&aftership_genkit.AfterShip{
			BaseURL: opt.AfterShipLLMHost,
		}),
		genkit.WithDefaultModel("aftership/claude-3-5-sonnet"),
	)

	if err != nil {
		panic(err)
	}
	return client
}

func (service *Agent) initAgentFlows() {
	genkit.DefineStreamingFlow[agentmodel.ProductListingInput, string, string](
		service.GenkitClient, "product_listings_simple", service.SimpleRun)
}

func (service *Agent) SimpleRun(ctx context.Context, input agentmodel.ProductListingInput, streamCallback core.StreamCallback[string]) (string, error) {
	log.L(ctx).Debug("product_listings_simple flow start", zap.Int("message_count", len(input.Messages)))

	var messages []*ai.Message
	mainPrompt, renderErr := service.promptLoader.RenderPrompt(ctx, "product_listings/system_main_render_card.prompt", nil)
	if renderErr != nil {
		log.L(ctx).Error("Failed to render main system prompt", zap.Error(renderErr))
	}
	mainPromptMessages, err := prompt.ConvertMessageToGenkitMessage(mainPrompt.Messages)
	if err != nil {
		return "", err
	}
	messages = append(messages, mainPromptMessages...)

	// Convert input messages to genkit messages
	for _, msg := range input.Messages {
		// Skip system messages
		if msg.Role == ai.RoleSystem {
			continue
		}
		var msgRole = msg.Role
		if msgRole == "assistant" {
			msgRole = ai.RoleModel
		}
		messages = append(messages, &ai.Message{
			Role:    msgRole,
			Content: []*ai.Part{ai.NewTextPart(msg.Content)},
		})
	}

	var modelStreamCallback ai.ModelStreamCallback
	if streamCallback != nil {
		modelStreamCallback = func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			if chunk.Content == nil {
				return nil // No content in this chunk
			}

			for _, part := range chunk.Content {
				if part.IsText() {
					return streamCallback(ctx, part.Text)
				}
			}
			return nil
		}
	}

	// Generate response using genkit with system prompt and tools
	resp, err := genkit.Generate(ctx, service.GenkitClient,
		ai.WithStreaming(modelStreamCallback),
		ai.WithMessages(messages...),
		ai.WithTools(service.activeTools...),
		ai.WithToolChoice(ai.ToolChoiceAuto),
	)

	if err != nil {
		return "", errors.WithCode(code.ErrUnknown, "Failed to generate response: %v", err)
	}

	// Return the generated response
	var responseContent string
	if resp != nil {
		responseContent = resp.Text()
	}

	return responseContent, nil
}
