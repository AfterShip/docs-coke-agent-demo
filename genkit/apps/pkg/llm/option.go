package llm

type Option struct {
	PromptDirectory  string `json:"prompt_directory" mapstructure:"prompt_directory" validate:"omitempty"`
	AfterShipLLMHost string `json:"aftership_llm_host" mapstructure:"aftership_llm_host" validate:"omitempty"`
}

func NewOption() *Option {
	return &Option{
		PromptDirectory:  "prompts",
		AfterShipLLMHost: "http://testing-incy-data-aigc.as-in.io",
	}
}
