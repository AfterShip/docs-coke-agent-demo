package aftership_genkit

import (
	"github.com/firebase/genkit/go/ai"
)

const provider = "aftership"

// Multimodal defines model capabilities for multimodal models
var Multimodal = ai.ModelSupports{
	Multiturn:  true,
	Tools:      true,
	SystemRole: true,
	Media:      true,
}

// supported anthropic models
var anthropicModels = map[string]ai.ModelInfo{
	"claude-4-sonnet": {
		Label:    "AfterShip Claude Sonnet 4",
		Supports: &Multimodal,
		Versions: []string{"us.anthropic.claude-sonnet-4-20250514-v1:0"},
	},
	"claude-3-5-sonnet": {
		Label:    "AfterShip Claude Sonnet 4",
		Supports: &Multimodal,
		Versions: []string{"anthropic.claude-3-5-sonnet-20241022-v2:0"},
	},
}
