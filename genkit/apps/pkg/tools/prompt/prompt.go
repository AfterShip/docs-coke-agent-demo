package prompt

import (
	"context"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/dotprompt/go/dotprompt"
	"github.com/mingyuans/errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Loader struct {
	promptDirectory string
	dotPrompt       *dotprompt.Dotprompt
	promptFuncCache map[string]dotprompt.PromptFunction
	locker          sync.Locker
}

func NewPromptLoader(promptDirectory string, options *dotprompt.DotpromptOptions) *Loader {
	return &Loader{
		promptDirectory: promptDirectory,
		dotPrompt:       dotprompt.NewDotprompt(options),
		promptFuncCache: make(map[string]dotprompt.PromptFunction),
		locker:          &sync.Mutex{},
	}
}

func (l *Loader) GetPromptFunc(ctx context.Context, filename string) (dotprompt.PromptFunction, error) {
	promptKey := strings.TrimSuffix(filename, ".prompt")
	if promptFunc, exists := l.promptFuncCache[promptKey]; exists {
		log.L(ctx).Debug("Using cached prompt", zap.String("promptKey", promptKey))
		return promptFunc, nil
	}

	l.locker.Lock()
	defer l.locker.Unlock()

	sourceFile := filepath.Join(l.promptDirectory, filename)
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read prompt file %s", sourceFile)
	}
	promptFunc, err := l.dotPrompt.Compile(string(source), nil)
	l.promptFuncCache[promptKey] = promptFunc
	return promptFunc, nil
}

func (l *Loader) RenderPrompt(ctx context.Context, filename string, input map[string]any) (dotprompt.RenderedPrompt, error) {
	promptFunc, err := l.GetPromptFunc(ctx, filename)
	if err != nil {
		return dotprompt.RenderedPrompt{}, err
	}

	arg := &dotprompt.DataArgument{
		Input: input,
	}
	return promptFunc(arg, nil)
}

func ConvertMessageToGenkitMessage(messages []dotprompt.Message) ([]*ai.Message, error) {
	convertedMessages := make([]*ai.Message, 0, len(messages))
	for _, message := range messages {
		parts := make([]*ai.Part, 0, len(message.Content))
		for _, content := range message.Content {
			convertedPart, err := convertPart(content)
			if err != nil {
				return nil, err
			}
			parts = append(parts, convertedPart)
		}

		if len(parts) > 0 {
			convertedMessages = append(convertedMessages, &ai.Message{
				Role:    ai.Role(message.Role),
				Content: parts,
			})
		}
	}
	return convertedMessages, nil
}

func convertPart(part dotprompt.Part) (*ai.Part, error) {
	switch p := part.(type) {
	case *dotprompt.TextPart:
		return ai.NewTextPart(p.Text), nil
	case *dotprompt.MediaPart:
		return ai.NewMediaPart(p.Media.ContentType, p.Media.URL), nil
	default:
		return nil, errors.Errorf("unknown part type: %T", p)
	}
}
