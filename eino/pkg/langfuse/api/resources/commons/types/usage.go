package types

// Usage represents token usage and cost information for LLM operations
type Usage struct {
	// Input token count
	Input *int `json:"input,omitempty"`

	// Output token count
	Output *int `json:"output,omitempty"`

	// Total token count (input + output)
	Total *int `json:"total,omitempty"`

	// Unit for token counting (e.g., "TOKENS", "CHARACTERS", "MILLISECONDS")
	Unit *string `json:"unit,omitempty"`

	// Cost information
	InputCost  *float64 `json:"inputCost,omitempty"`
	OutputCost *float64 `json:"outputCost,omitempty"`
	TotalCost  *float64 `json:"totalCost,omitempty"`
}

// UsageCreateRequest represents a request structure for usage data
type UsageCreateRequest struct {
	// Input token count
	Input *int `json:"input,omitempty"`

	// Output token count
	Output *int `json:"output,omitempty"`

	// Total token count (input + output)
	Total *int `json:"total,omitempty"`

	// Unit for token counting (e.g., "TOKENS", "CHARACTERS", "MILLISECONDS")
	Unit *string `json:"unit,omitempty"`

	// Cost information
	InputCost  *float64 `json:"inputCost,omitempty"`
	OutputCost *float64 `json:"outputCost,omitempty"`
	TotalCost  *float64 `json:"totalCost,omitempty"`
}

// CalculateTotalTokens calculates total tokens from input and output if not already set
func (u *Usage) CalculateTotalTokens() {
	if u.Total == nil && u.Input != nil && u.Output != nil {
		total := *u.Input + *u.Output
		u.Total = &total
	}
}

// CalculateTotalCost calculates total cost from input and output costs if not already set
func (u *Usage) CalculateTotalCost() {
	if u.TotalCost == nil && u.InputCost != nil && u.OutputCost != nil {
		total := *u.InputCost + *u.OutputCost
		u.TotalCost = &total
	}
}

// NewUsage creates a new Usage instance with basic token counts
func NewUsage(inputTokens, outputTokens int) *Usage {
	total := inputTokens + outputTokens
	unit := "TOKENS"

	return &Usage{
		Input:  &inputTokens,
		Output: &outputTokens,
		Total:  &total,
		Unit:   &unit,
	}
}

// NewUsageWithCost creates a new Usage instance with token counts and cost information
func NewUsageWithCost(inputTokens, outputTokens int, inputCost, outputCost float64) *Usage {
	total := inputTokens + outputTokens
	totalCost := inputCost + outputCost
	unit := "TOKENS"

	return &Usage{
		Input:      &inputTokens,
		Output:     &outputTokens,
		Total:      &total,
		Unit:       &unit,
		InputCost:  &inputCost,
		OutputCost: &outputCost,
		TotalCost:  &totalCost,
	}
}
