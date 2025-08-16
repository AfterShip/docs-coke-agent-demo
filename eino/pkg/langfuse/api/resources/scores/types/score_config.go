package types

import (
	"fmt"

	commonTypes "eino/pkg/langfuse/api/resources/commons/types"
)

// ScoreConfig represents a score configuration
type ScoreConfig struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	DataType    commonTypes.ScoreDataType `json:"dataType"`
	IsArchived  bool                      `json:"isArchived"`
	Description *string                   `json:"description,omitempty"`
	Categories  []ScoreCategory           `json:"categories,omitempty"`
	Range       *ScoreRange               `json:"range,omitempty"`
}

// ScoreCategory represents a category for categorical scores
type ScoreCategory struct {
	Value       string  `json:"value"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"`
}

// ScoreRange represents a range for numeric scores
type ScoreRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// GetScoreConfigsRequest represents a request to get score configurations
type GetScoreConfigsRequest struct {
	ProjectID  string `json:"projectId,omitempty"`
	Page       *int   `json:"page,omitempty"`
	Limit      *int   `json:"limit,omitempty"`
	DataType   *commonTypes.ScoreDataType `json:"dataType,omitempty"`
	IsArchived *bool  `json:"isArchived,omitempty"`
}

// CreateScoreConfigRequest represents a request to create a score configuration
type CreateScoreConfigRequest struct {
	Name        string                    `json:"name"`
	DataType    commonTypes.ScoreDataType `json:"dataType"`
	Description *string                   `json:"description,omitempty"`
	Categories  []ScoreCategory           `json:"categories,omitempty"`
	Range       *ScoreRange               `json:"range,omitempty"`
}

// UpdateScoreConfigRequest represents a request to update a score configuration
type UpdateScoreConfigRequest struct {
	ID          string                     `json:"id"`
	Name        *string                    `json:"name,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Categories  []ScoreCategory            `json:"categories,omitempty"`
	Range       *ScoreRange                `json:"range,omitempty"`
	IsArchived  *bool                      `json:"isArchived,omitempty"`
}

// Validate validates the create score config request
func (req *CreateScoreConfigRequest) Validate() error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	
	// Validate categories for categorical scores
	if req.DataType == commonTypes.ScoreDataTypeCategorical {
		if len(req.Categories) == 0 {
			return &ValidationError{Field: "categories", Message: "categories are required for categorical scores"}
		}
		
		for i, category := range req.Categories {
			if category.Value == "" {
				return &ValidationError{Field: "categories", Message: fmt.Sprintf("category at index %d must have a value", i)}
			}
			if category.Label == "" {
				return &ValidationError{Field: "categories", Message: fmt.Sprintf("category at index %d must have a label", i)}
			}
		}
	}
	
	// Validate range for numeric scores
	if req.DataType == commonTypes.ScoreDataTypeNumeric && req.Range != nil {
		if req.Range.Min != nil && req.Range.Max != nil && *req.Range.Min >= *req.Range.Max {
			return &ValidationError{Field: "range", Message: "min value must be less than max value"}
		}
	}
	
	return nil
}

// Validate validates the update score config request
func (req *UpdateScoreConfigRequest) Validate() error {
	if req.ID == "" {
		return &ValidationError{Field: "id", Message: "id is required"}
	}
	
	// Validate categories if provided
	if len(req.Categories) > 0 {
		for i, category := range req.Categories {
			if category.Value == "" {
				return &ValidationError{Field: "categories", Message: fmt.Sprintf("category at index %d must have a value", i)}
			}
			if category.Label == "" {
				return &ValidationError{Field: "categories", Message: fmt.Sprintf("category at index %d must have a label", i)}
			}
		}
	}
	
	// Validate range if provided
	if req.Range != nil && req.Range.Min != nil && req.Range.Max != nil && *req.Range.Min >= *req.Range.Max {
		return &ValidationError{Field: "range", Message: "min value must be less than max value"}
	}
	
	return nil
}

// IsNumeric returns true if this is a numeric score config
func (sc *ScoreConfig) IsNumeric() bool {
	return sc.DataType == commonTypes.ScoreDataTypeNumeric
}

// IsCategorical returns true if this is a categorical score config
func (sc *ScoreConfig) IsCategorical() bool {
	return sc.DataType == commonTypes.ScoreDataTypeCategorical
}

// IsBoolean returns true if this is a boolean score config
func (sc *ScoreConfig) IsBoolean() bool {
	return sc.DataType == commonTypes.ScoreDataTypeBoolean
}

// GetCategoryLabels returns a list of category labels
func (sc *ScoreConfig) GetCategoryLabels() []string {
	if !sc.IsCategorical() {
		return nil
	}
	
	labels := make([]string, len(sc.Categories))
	for i, category := range sc.Categories {
		labels[i] = category.Label
	}
	return labels
}

// GetCategoryValues returns a list of category values
func (sc *ScoreConfig) GetCategoryValues() []string {
	if !sc.IsCategorical() {
		return nil
	}
	
	values := make([]string, len(sc.Categories))
	for i, category := range sc.Categories {
		values[i] = category.Value
	}
	return values
}

// IsValidValue checks if a value is valid for this score config
func (sc *ScoreConfig) IsValidValue(value interface{}) bool {
	switch sc.DataType {
	case commonTypes.ScoreDataTypeNumeric:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			if sc.Range != nil {
				var floatVal float64
				switch val := v.(type) {
				case float64:
					floatVal = val
				case float32:
					floatVal = float64(val)
				case int:
					floatVal = float64(val)
				case int64:
					floatVal = float64(val)
				// Add more type conversions as needed
				default:
					return false
				}
				
				if sc.Range.Min != nil && floatVal < *sc.Range.Min {
					return false
				}
				if sc.Range.Max != nil && floatVal > *sc.Range.Max {
					return false
				}
			}
			return true
		default:
			return false
		}
		
	case commonTypes.ScoreDataTypeBoolean:
		_, ok := value.(bool)
		return ok
		
	case commonTypes.ScoreDataTypeCategorical:
		strVal, ok := value.(string)
		if !ok {
			return false
		}
		
		for _, category := range sc.Categories {
			if category.Value == strVal {
				return true
			}
		}
		return false
		
	default:
		return false
	}
}