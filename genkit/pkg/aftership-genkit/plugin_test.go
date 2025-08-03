package aftership_genkit

import (
	"context"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/logger"
	"github.com/firebase/genkit/go/genkit"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func Test_init(t *testing.T) {
	ctx := context.Background()
	logger.SetLevel(slog.LevelDebug)

	//Please set the environment variables before running the test
	assert.NotEmpty(t, os.Getenv("AFTERSHIP_API_KEY"))
	assert.NotEmpty(t, os.Getenv("AFTERSHIP_LLM_BASE_URL"))

	// Initialize Genkit with the Anthropic plugin and Claude model.
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&AfterShip{}),
		genkit.WithDefaultModel("aftership/claude-sonnet-4"),
	)

	resp, err := genkit.Generate(ctx, g,
		ai.WithPrompt("Hi"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
