package aftership_genkit

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/invopop/jsonschema"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	MaxNumberOfTokens = 8192
	ToolNameRegex     = `^[a-zA-Z0-9_-]{1,64}$`
)

type AfterShip struct {
	APIKey  string
	BaseURL string

	client  *anthropic.Client
	mu      sync.Mutex
	initted bool
}

func (a *AfterShip) Name() string {
	return provider
}

func (a *AfterShip) Init(ctx context.Context, g *genkit.Genkit) (err error) {
	if a == nil {
		a = &AfterShip{}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.initted {
		return errors.New("plugin already initialized")
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("AfterShip.Init: %w", err)
		}
	}()

	apiKey := a.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("AM_API_KEY")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required. Set APIKey field or AM_API_KEY environment variable")
	}

	baseURL := a.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("AFTERSHIP_LLM_BASE_URL")
	}
	if baseURL == "" {
		return fmt.Errorf("API key is required. Set APIKey field or AFTERSHIP_LLM_BASE_URL environment variable")
	}

	c := anthropic.NewClient(
		option.WithHeaderAdd("am-api-key", apiKey),
		//option.WithMiddleware(streamCorrectionMiddleware),
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	a.initted = true
	a.client = &c

	for name, mi := range anthropicModels {
		defineAnthropicModel(g, a.client, name, mi)
	}

	return nil
}

// DefineModel adds the model to the registry
func (a *AfterShip) DefineModel(g *genkit.Genkit, name string, info *ai.ModelInfo) (ai.Model, error) {
	var mi ai.ModelInfo
	if info == nil {
		var ok bool
		mi, ok = anthropicModels[name]
		if !ok {
			return nil, fmt.Errorf("%s.DefineModel: called with unknown model %q and nil ModelInfo", provider, name)
		}
	} else {
		mi = *info
	}
	return defineAnthropicModel(g, a.client, name, mi), nil
}

func defineAnthropicModel(g *genkit.Genkit, client *anthropic.Client, name string, info ai.ModelInfo) ai.Model {
	// First, try to find an existing model
	if existing := genkit.LookupModel(g, provider, name); existing != nil {
		return existing
	}

	meta := &ai.ModelInfo{
		Label:    provider + "-" + name,
		Supports: info.Supports,
		Versions: info.Versions,
	}
	return genkit.DefineModel(g, provider, name, meta, func(
		ctx context.Context,
		input *ai.ModelRequest,
		cb func(context.Context, *ai.ModelResponseChunk) error,
	) (*ai.ModelResponse, error) {
		return anthropicGenerate(ctx, client, name, input, cb)
	})
}

// generate function defines how a generate request is done in AfterShip models
func anthropicGenerate(
	ctx context.Context,
	client *anthropic.Client,
	model string,
	input *ai.ModelRequest,
	cb func(context.Context, *ai.ModelResponseChunk) error,
) (*ai.ModelResponse, error) {
	req, err := toAnthropicRequest(model, input)
	if err != nil {
		return nil, fmt.Errorf("unable to generate anthropic request: %w", err)
	}

	// no streaming
	if cb == nil {
		msg, err := client.Messages.New(ctx, *req)
		if err != nil {
			return nil, err
		}

		r, err := anthropicToGenkitResponse(msg)
		if err != nil {
			return nil, err
		}

		r.Request = input
		return r, nil
	} else {
		stream := client.Messages.NewStreaming(ctx, *req)
		message := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			fmt.Printf("Received event: %s\n", event.Type)
			fmt.Printf("%s\n", event.RawJSON())

			err := message.Accumulate(event)
			if err != nil {
				return nil, err
			}

			switch event := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				cb(ctx, &ai.ModelResponseChunk{
					Content: []*ai.Part{
						{
							Text: event.Delta.Text,
						},
					},
				})
			case anthropic.MessageStopEvent:
				r, err := anthropicToGenkitResponse(&message)
				if err != nil {
					return nil, err
				}
				r.Request = input
				return r, nil
			}
		}
		if stream.Err() != nil {
			return nil, stream.Err()
		}
	}

	return nil, nil
}

func toAnthropicRole(role ai.Role) (anthropic.MessageParamRole, error) {
	switch role {
	case ai.RoleUser:
		return anthropic.MessageParamRoleUser, nil
	case ai.RoleModel:
		return anthropic.MessageParamRoleAssistant, nil
	case ai.RoleTool:
		return anthropic.MessageParamRoleAssistant, nil
	default:
		return "", fmt.Errorf("unknown role given: %q", role)
	}
}

// toAnthropicRequest translates [ai.ModelRequest] to an AfterShip request
func toAnthropicRequest(model string, i *ai.ModelRequest) (*anthropic.MessageNewParams, error) {
	messages := make([]anthropic.MessageParam, 0)

	c, err := configFromRequest(i)
	if err != nil {
		return nil, err
	}

	// minimum required data to perform a request
	req := anthropic.MessageNewParams{}

	// Use default version (if version info exists in model definition)
	if modelInfo, exists := anthropicModels[model]; exists && len(modelInfo.Versions) > 0 {
		req.Model = anthropic.Model(modelInfo.Versions[0]) // Use first version as default
	} else {
		req.Model = anthropic.Model(model) // Fallback to using model name
	}
	req.MaxTokens = int64(MaxNumberOfTokens)

	if c.MaxOutputTokens != 0 {
		req.MaxTokens = int64(c.MaxOutputTokens)
	}
	if c.Version != "" {
		req.Model = anthropic.Model(c.Version) // User-specified version has highest priority
	}
	if c.Temperature != 0 {
		req.Temperature = anthropic.Float(c.Temperature)
	}
	if c.TopK != 0 {
		req.TopK = anthropic.Int(int64(c.TopK))
	}
	if c.TopP != 0 {
		req.TopP = anthropic.Float(float64(c.TopP))
	}
	if len(c.StopSequences) > 0 {
		req.StopSequences = c.StopSequences
	}

	// configure system prompt (if given)
	sysBlocks := []anthropic.TextBlockParam{}
	for _, message := range i.Messages {
		if message.Role == ai.RoleSystem {
			// only text is supported for system messages
			sysBlocks = append(sysBlocks, anthropic.TextBlockParam{Text: message.Text()})
		} else if message.Content[len(message.Content)-1].IsToolResponse() {
			// if the last message is a ToolResponse, the conversation must continue
			// and the ToolResponse message must be sent as a user
			// see: https://docs.anthropic.com/en/docs/build-with-claude/tool-use#handling-tool-use-and-tool-result-content-blocks
			parts, err := toAnthropicParts(message.Content)
			if err != nil {
				return nil, err
			}
			messages = append(messages, anthropic.NewUserMessage(parts...))
		} else {
			parts, err := toAnthropicParts(message.Content)
			if err != nil {
				return nil, err
			}
			role, err := toAnthropicRole(message.Role)
			if err != nil {
				return nil, err
			}
			messages = append(messages, anthropic.MessageParam{
				Role:    role,
				Content: parts,
			})
		}
	}

	req.System = sysBlocks
	req.Messages = messages

	tools, err := toAnthropicTools(i.Tools)
	if err != nil {
		return nil, err
	}
	req.Tools = tools

	return &req, nil
}

// mapToStruct unmarshals a map[String]any to the expected type
func mapToStruct(m map[string]any, v any) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, v)
}

// configFromRequest converts any supported config type to [ai.GenerationCommonConfig]
func configFromRequest(input *ai.ModelRequest) (*ai.GenerationCommonConfig, error) {
	var result ai.GenerationCommonConfig

	switch config := input.Config.(type) {
	case ai.GenerationCommonConfig:
		result = config
	case *ai.GenerationCommonConfig:
		result = *config
	case map[string]any:
		if err := mapToStruct(config, &result); err != nil {
			return nil, err
		}
	case nil:
		// Empty configuration is considered valid
	default:
		return nil, fmt.Errorf("unexpected config type: %T", input.Config)
	}
	return &result, nil
}

// toAnthropicTools translates [ai.ToolDefinition] to an anthropic.ToolParam type
func toAnthropicTools(tools []*ai.ToolDefinition) ([]anthropic.ToolUnionParam, error) {
	resp := make([]anthropic.ToolUnionParam, 0)
	regex := regexp.MustCompile(ToolNameRegex)

	for _, t := range tools {
		if t.Name == "" {
			return nil, fmt.Errorf("tool name is required")
		}
		if !regex.MatchString(t.Name) {
			return nil, fmt.Errorf("tool name must match regex: %s", ToolNameRegex)
		}

		resp = append(resp, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: toAnthropicSchema[map[string]any](),
			},
		})
	}

	return resp, nil
}

// toAnthropicSchema generates a JSON schema for the requested input type
func toAnthropicSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: true,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

// toAnthropicParts translates [ai.Part] to an anthropic.ContentBlockParamUnion type
func toAnthropicParts(parts []*ai.Part) ([]anthropic.ContentBlockParamUnion, error) {
	blocks := []anthropic.ContentBlockParamUnion{}

	for _, p := range parts {
		switch {
		case p.IsText():
			blocks = append(blocks, anthropic.NewTextBlock(p.Text))
		case p.IsMedia():
			contentType, data, _ := Data(p)
			blocks = append(blocks, anthropic.NewImageBlockBase64(contentType, base64.StdEncoding.EncodeToString(data)))
		case p.IsData():
			contentType, data, _ := Data(p)
			blocks = append(blocks, anthropic.NewImageBlockBase64(contentType, base64.RawStdEncoding.EncodeToString(data)))
		case p.IsToolRequest():
			toolReq := p.ToolRequest
			blocks = append(blocks, anthropic.NewToolUseBlock(toolReq.Ref, toolReq.Input, toolReq.Name))
		case p.IsToolResponse():
			toolResp := p.ToolResponse
			output, err := json.Marshal(toolResp.Output)
			if err != nil {
				return nil, fmt.Errorf("unable to parse tool response, err: %w", err)
			}
			blocks = append(blocks, anthropic.NewToolResultBlock(toolResp.Ref, string(output), false))
		default:
			return nil, errors.New("unknown part type in the request")
		}
	}

	return blocks, nil
}

// anthropicToGenkitResponse translates an AfterShip Message to [ai.ModelResponse]
func anthropicToGenkitResponse(m *anthropic.Message) (*ai.ModelResponse, error) {
	r := ai.ModelResponse{}

	switch m.StopReason {
	case anthropic.StopReasonMaxTokens:
		r.FinishReason = ai.FinishReasonLength
	case anthropic.StopReasonStopSequence:
		r.FinishReason = ai.FinishReasonStop
	case anthropic.StopReasonEndTurn:
		r.FinishReason = ai.FinishReasonStop
	case anthropic.StopReasonToolUse:
		r.FinishReason = ai.FinishReasonStop
	default:
		r.FinishReason = ai.FinishReasonUnknown
	}

	msg := &ai.Message{}
	msg.Role = ai.RoleModel
	for _, part := range m.Content {
		var p *ai.Part
		switch part.AsAny().(type) {
		case anthropic.TextBlock:
			p = ai.NewTextPart(string(part.Text))
		case anthropic.ToolUseBlock:
			p = ai.NewToolRequestPart(&ai.ToolRequest{
				Ref:   part.ID,
				Input: part.Input,
				Name:  part.Name,
			})
		default:
			return nil, fmt.Errorf("unknown part: %#v", part)
		}
		msg.Content = append(msg.Content, p)
	}

	r.Message = msg
	r.Usage = &ai.GenerationUsage{
		InputTokens:  int(m.Usage.InputTokens),
		OutputTokens: int(m.Usage.OutputTokens),
	}
	return &r, nil
}
