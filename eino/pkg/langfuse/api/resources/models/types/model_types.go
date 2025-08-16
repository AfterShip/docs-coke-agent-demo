package types

import (
	"time"

	"eino/pkg/langfuse/api/resources/utils/pagination/types"
)

// Model represents a language model configuration
type Model struct {
	// Unique identifier for the model
	ID string `json:"id"`

	// Name of the model
	ModelName string `json:"modelName"`

	// Match pattern for identifying the model
	MatchPattern string `json:"matchPattern"`

	// Timestamp when the model was created
	StartDate time.Time `json:"startDate"`

	// Timestamp when the model was deprecated (optional)
	EndDate *time.Time `json:"endDate,omitempty"`

	// Project ID this model belongs to
	ProjectID string `json:"projectId"`

	// Input price per token
	InputPrice *float64 `json:"inputPrice,omitempty"`

	// Output price per token
	OutputPrice *float64 `json:"outputPrice,omitempty"`

	// Total price per request (for models that don't charge per token)
	TotalPrice *float64 `json:"totalPrice,omitempty"`

	// Currency for pricing
	Currency string `json:"currency"`

	// Unit for pricing (e.g., "TOKENS", "REQUESTS", "CHARACTERS")
	Unit ModelPricingUnit `json:"unit"`

	// Tokenizer ID for token counting
	TokenizerID *string `json:"tokenizerId,omitempty"`

	// Model configuration
	Config *ModelConfig `json:"config,omitempty"`

	// Timestamp when the model was created
	CreatedAt time.Time `json:"createdAt"`

	// Timestamp when the model was last updated
	UpdatedAt time.Time `json:"updatedAt"`

	// User who created the model
	CreatedBy *string `json:"createdBy,omitempty"`
}

// ModelPricingUnit represents the unit for model pricing
type ModelPricingUnit string

const (
	ModelPricingUnitTokens     ModelPricingUnit = "TOKENS"
	ModelPricingUnitRequests   ModelPricingUnit = "REQUESTS"
	ModelPricingUnitCharacters ModelPricingUnit = "CHARACTERS"
	ModelPricingUnitSeconds    ModelPricingUnit = "SECONDS"
)

// ModelConfig represents configuration for a model
type ModelConfig struct {
	// Provider of the model
	Provider string `json:"provider"`

	// Model family (e.g., "gpt", "claude", "llama")
	ModelFamily *string `json:"modelFamily,omitempty"`

	// Maximum context length
	ContextLength *int `json:"contextLength,omitempty"`

	// Maximum output tokens
	MaxOutputTokens *int `json:"maxOutputTokens,omitempty"`

	// Default temperature
	DefaultTemperature *float64 `json:"defaultTemperature,omitempty"`

	// Supported features
	Features []ModelFeature `json:"features,omitempty"`

	// Additional configuration
	AdditionalConfig map[string]interface{} `json:"additionalConfig,omitempty"`
}

// ModelFeature represents features supported by a model
type ModelFeature string

const (
	ModelFeatureChat         ModelFeature = "chat"
	ModelFeatureCompletion   ModelFeature = "completion"
	ModelFeatureEmbedding    ModelFeature = "embedding"
	ModelFeatureFunctionCall ModelFeature = "function_calling"
	ModelFeatureVision       ModelFeature = "vision"
	ModelFeatureStreaming    ModelFeature = "streaming"
)

// GetModelsRequest represents a request to list models
type GetModelsRequest struct {
	ProjectID    string     `json:"projectId,omitempty"`
	Page         *int       `json:"page,omitempty"`
	Limit        *int       `json:"limit,omitempty"`
	ModelName    *string    `json:"modelName,omitempty"`
	Provider     *string    `json:"provider,omitempty"`
	ModelFamily  *string    `json:"modelFamily,omitempty"`
	Unit         *ModelPricingUnit `json:"unit,omitempty"`
	FromDate     *time.Time `json:"fromDate,omitempty"`
	ToDate       *time.Time `json:"toDate,omitempty"`
	IncludeDeprecated *bool `json:"includeDeprecated,omitempty"`
}

// GetModelsResponse represents the response from listing models
type GetModelsResponse struct {
	Data []Model              `json:"data"`
	Meta types.MetaResponse   `json:"meta"`
}

// CreateModelRequest represents a request to create a model
type CreateModelRequest struct {
	ModelName     string           `json:"modelName"`
	MatchPattern  string           `json:"matchPattern"`
	StartDate     time.Time        `json:"startDate"`
	EndDate       *time.Time       `json:"endDate,omitempty"`
	InputPrice    *float64         `json:"inputPrice,omitempty"`
	OutputPrice   *float64         `json:"outputPrice,omitempty"`
	TotalPrice    *float64         `json:"totalPrice,omitempty"`
	Currency      string           `json:"currency"`
	Unit          ModelPricingUnit `json:"unit"`
	TokenizerID   *string          `json:"tokenizerId,omitempty"`
	Config        *ModelConfig     `json:"config,omitempty"`
}

// CreateModelResponse represents the response from creating a model
type CreateModelResponse struct {
	ID            string           `json:"id"`
	ModelName     string           `json:"modelName"`
	MatchPattern  string           `json:"matchPattern"`
	StartDate     time.Time        `json:"startDate"`
	EndDate       *time.Time       `json:"endDate,omitempty"`
	ProjectID     string           `json:"projectId"`
	InputPrice    *float64         `json:"inputPrice,omitempty"`
	OutputPrice   *float64         `json:"outputPrice,omitempty"`
	TotalPrice    *float64         `json:"totalPrice,omitempty"`
	Currency      string           `json:"currency"`
	Unit          ModelPricingUnit `json:"unit"`
	TokenizerID   *string          `json:"tokenizerId,omitempty"`
	Config        *ModelConfig     `json:"config,omitempty"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
	CreatedBy     *string          `json:"createdBy,omitempty"`
}

// UpdateModelRequest represents a request to update a model
type UpdateModelRequest struct {
	ModelName     *string          `json:"modelName,omitempty"`
	MatchPattern  *string          `json:"matchPattern,omitempty"`
	StartDate     *time.Time       `json:"startDate,omitempty"`
	EndDate       *time.Time       `json:"endDate,omitempty"`
	InputPrice    *float64         `json:"inputPrice,omitempty"`
	OutputPrice   *float64         `json:"outputPrice,omitempty"`
	TotalPrice    *float64         `json:"totalPrice,omitempty"`
	Currency      *string          `json:"currency,omitempty"`
	Unit          *ModelPricingUnit `json:"unit,omitempty"`
	TokenizerID   *string          `json:"tokenizerId,omitempty"`
	Config        *ModelConfig     `json:"config,omitempty"`
}

// ModelUsageStats represents usage statistics for models
type ModelUsageStats struct {
	TotalModels       int                    `json:"totalModels"`
	ActiveModels      int                    `json:"activeModels"`
	DeprecatedModels  int                    `json:"deprecatedModels"`
	ModelsByProvider  map[string]int         `json:"modelsByProvider"`
	ModelsByUnit      map[string]int         `json:"modelsByUnit"`
	UsageByModel      map[string]ModelUsage  `json:"usageByModel"`
	CostByModel       map[string]float64     `json:"costByModel"`
	DateRange         *DateRange             `json:"dateRange,omitempty"`
}

// ModelUsage represents usage statistics for a specific model
type ModelUsage struct {
	ModelID        string    `json:"modelId"`
	ModelName      string    `json:"modelName"`
	TotalRequests  int       `json:"totalRequests"`
	TotalTokens    *int      `json:"totalTokens,omitempty"`
	InputTokens    *int      `json:"inputTokens,omitempty"`
	OutputTokens   *int      `json:"outputTokens,omitempty"`
	TotalCost      float64   `json:"totalCost"`
	AverageCost    float64   `json:"averageCost"`
	LastUsed       time.Time `json:"lastUsed"`
}

// DateRange represents a date range
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// GetModelUsageStatsRequest represents a request to get model usage statistics
type GetModelUsageStatsRequest struct {
	ProjectID     string     `json:"projectId,omitempty"`
	ModelID       *string    `json:"modelId,omitempty"`
	ModelName     *string    `json:"modelName,omitempty"`
	Provider      *string    `json:"provider,omitempty"`
	FromTimestamp *time.Time `json:"fromTimestamp,omitempty"`
	ToTimestamp   *time.Time `json:"toTimestamp,omitempty"`
}

// ModelMatchRequest represents a request to match a model name to a configured model
type ModelMatchRequest struct {
	ModelName string `json:"modelName"`
}

// ModelMatchResponse represents the response from matching a model
type ModelMatchResponse struct {
	Model     *Model `json:"model"`
	Matched   bool   `json:"matched"`
	MatchType string `json:"matchType"` // "exact", "pattern", "fallback"
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validate validates the GetModelsRequest
func (req *GetModelsRequest) Validate() error {
	if req.Limit != nil && (*req.Limit < 1 || *req.Limit > 1000) {
		return &ValidationError{Field: "limit", Message: "limit must be between 1 and 1000"}
	}
	
	if req.Page != nil && *req.Page < 1 {
		return &ValidationError{Field: "page", Message: "page must be greater than 0"}
	}
	
	if req.FromDate != nil && req.ToDate != nil && req.FromDate.After(*req.ToDate) {
		return &ValidationError{Field: "dates", Message: "fromDate cannot be after toDate"}
	}
	
	if req.Unit != nil {
		validUnits := map[ModelPricingUnit]bool{
			ModelPricingUnitTokens:     true,
			ModelPricingUnitRequests:   true,
			ModelPricingUnitCharacters: true,
			ModelPricingUnitSeconds:    true,
		}
		if !validUnits[*req.Unit] {
			return &ValidationError{Field: "unit", Message: "invalid pricing unit"}
		}
	}
	
	return nil
}

// Validate validates the CreateModelRequest
func (req *CreateModelRequest) Validate() error {
	if req.ModelName == "" {
		return &ValidationError{Field: "modelName", Message: "modelName is required"}
	}
	
	if len(req.ModelName) > 255 {
		return &ValidationError{Field: "modelName", Message: "modelName must be 255 characters or less"}
	}
	
	if req.MatchPattern == "" {
		return &ValidationError{Field: "matchPattern", Message: "matchPattern is required"}
	}
	
	if req.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	
	if len(req.Currency) != 3 {
		return &ValidationError{Field: "currency", Message: "currency must be a 3-letter code (e.g., USD, EUR)"}
	}
	
	validUnits := map[ModelPricingUnit]bool{
		ModelPricingUnitTokens:     true,
		ModelPricingUnitRequests:   true,
		ModelPricingUnitCharacters: true,
		ModelPricingUnitSeconds:    true,
	}
	if !validUnits[req.Unit] {
		return &ValidationError{Field: "unit", Message: "invalid pricing unit"}
	}
	
	// Validate pricing based on unit
	switch req.Unit {
	case ModelPricingUnitTokens:
		if req.InputPrice == nil && req.OutputPrice == nil {
			return &ValidationError{Field: "pricing", Message: "inputPrice or outputPrice is required for token-based pricing"}
		}
		if req.TotalPrice != nil {
			return &ValidationError{Field: "totalPrice", Message: "totalPrice should not be set for token-based pricing"}
		}
	case ModelPricingUnitRequests, ModelPricingUnitCharacters, ModelPricingUnitSeconds:
		if req.TotalPrice == nil {
			return &ValidationError{Field: "totalPrice", Message: "totalPrice is required for non-token pricing"}
		}
		if req.InputPrice != nil || req.OutputPrice != nil {
			return &ValidationError{Field: "pricing", Message: "inputPrice/outputPrice should not be set for non-token pricing"}
		}
	}
	
	if req.InputPrice != nil && *req.InputPrice < 0 {
		return &ValidationError{Field: "inputPrice", Message: "inputPrice cannot be negative"}
	}
	
	if req.OutputPrice != nil && *req.OutputPrice < 0 {
		return &ValidationError{Field: "outputPrice", Message: "outputPrice cannot be negative"}
	}
	
	if req.TotalPrice != nil && *req.TotalPrice < 0 {
		return &ValidationError{Field: "totalPrice", Message: "totalPrice cannot be negative"}
	}
	
	if req.EndDate != nil && req.EndDate.Before(req.StartDate) {
		return &ValidationError{Field: "endDate", Message: "endDate cannot be before startDate"}
	}
	
	if req.Config != nil {
		if err := req.Config.Validate(); err != nil {
			return err
		}
	}
	
	return nil
}

// Validate validates the ModelMatchRequest
func (req *ModelMatchRequest) Validate() error {
	if req.ModelName == "" {
		return &ValidationError{Field: "modelName", Message: "modelName is required"}
	}
	
	return nil
}

// Validate validates the ModelConfig
func (config *ModelConfig) Validate() error {
	if config.Provider == "" {
		return &ValidationError{Field: "config.provider", Message: "provider is required"}
	}
	
	if config.ContextLength != nil && *config.ContextLength < 1 {
		return &ValidationError{Field: "config.contextLength", Message: "contextLength must be greater than 0"}
	}
	
	if config.MaxOutputTokens != nil && *config.MaxOutputTokens < 1 {
		return &ValidationError{Field: "config.maxOutputTokens", Message: "maxOutputTokens must be greater than 0"}
	}
	
	if config.DefaultTemperature != nil && (*config.DefaultTemperature < 0 || *config.DefaultTemperature > 2) {
		return &ValidationError{Field: "config.defaultTemperature", Message: "defaultTemperature must be between 0 and 2"}
	}
	
	// Validate features
	validFeatures := map[ModelFeature]bool{
		ModelFeatureChat:         true,
		ModelFeatureCompletion:   true,
		ModelFeatureEmbedding:    true,
		ModelFeatureFunctionCall: true,
		ModelFeatureVision:       true,
		ModelFeatureStreaming:    true,
	}
	
	for _, feature := range config.Features {
		if !validFeatures[feature] {
			return &ValidationError{Field: "config.features", Message: "invalid feature: " + string(feature)}
		}
	}
	
	return nil
}

// IsTokenBased returns true if the model uses token-based pricing
func (m *Model) IsTokenBased() bool {
	return m.Unit == ModelPricingUnitTokens
}

// IsActive returns true if the model is currently active
func (m *Model) IsActive() bool {
	return m.EndDate == nil || m.EndDate.After(time.Now())
}

// HasFeature returns true if the model supports the specified feature
func (m *Model) HasFeature(feature ModelFeature) bool {
	if m.Config == nil {
		return false
	}
	
	for _, f := range m.Config.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetInputCost calculates the input cost for a given number of units
func (m *Model) GetInputCost(units int) float64 {
	if m.InputPrice == nil {
		return 0
	}
	return *m.InputPrice * float64(units)
}

// GetOutputCost calculates the output cost for a given number of units
func (m *Model) GetOutputCost(units int) float64 {
	if m.OutputPrice == nil {
		return 0
	}
	return *m.OutputPrice * float64(units)
}

// GetTotalCost calculates the total cost for a single request
func (m *Model) GetTotalCost() float64 {
	if m.TotalPrice == nil {
		return 0
	}
	return *m.TotalPrice
}

// NewTokenBasedModel creates a new token-based model request
func NewTokenBasedModel(name, matchPattern string, inputPrice, outputPrice float64, currency string) *CreateModelRequest {
	return &CreateModelRequest{
		ModelName:    name,
		MatchPattern: matchPattern,
		StartDate:    time.Now(),
		InputPrice:   &inputPrice,
		OutputPrice:  &outputPrice,
		Currency:     currency,
		Unit:         ModelPricingUnitTokens,
	}
}

// NewRequestBasedModel creates a new request-based model request
func NewRequestBasedModel(name, matchPattern string, totalPrice float64, currency string) *CreateModelRequest {
	return &CreateModelRequest{
		ModelName:    name,
		MatchPattern: matchPattern,
		StartDate:    time.Now(),
		TotalPrice:   &totalPrice,
		Currency:     currency,
		Unit:         ModelPricingUnitRequests,
	}
}