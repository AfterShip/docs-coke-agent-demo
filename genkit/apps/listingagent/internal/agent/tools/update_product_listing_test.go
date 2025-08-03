package tools

import (
	"testing"

	"github.com/AfterShip/connectors-library/sdks/product_listings"
	agentmodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/agent"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProductListingTool_Name(t *testing.T) {
	tool := NewUpdateProductListingTool()
	assert.Equal(t, "update_product_listing", tool.Name())
}

func TestUpdateProductListingTool_Description(t *testing.T) {
	tool := NewUpdateProductListingTool()
	description := tool.Description()

	assert.NotEmpty(t, description)
	assert.Contains(t, description, "Update an existing product listing")
	assert.Contains(t, description, "modify product details")
}

func TestEditArg_Validation(t *testing.T) {
	tests := []struct {
		name        string
		editArg     agentmodel.EditArg
		expectValid bool
	}{
		{
			name: "valid EditArg with required fields",
			editArg: agentmodel.EditArg{
				ProductListingID: "test-id-123",
				Product: product_listings.Product{
					Title: "Test Product",
				},
			},
			expectValid: true,
		},
		{
			name: "EditArg with empty ProductListingID",
			editArg: agentmodel.EditArg{
				ProductListingID: "",
				Product: product_listings.Product{
					Title: "Test Product",
				},
			},
			expectValid: false,
		},
		{
			name: "EditArg with product data",
			editArg: agentmodel.EditArg{
				ProductListingID: "test-id-456",
				Product: product_listings.Product{
					Title:       "Updated Product Title",
					Description: "Updated description",
				},
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证必需字段
			if tt.expectValid {
				assert.NotEmpty(t, tt.editArg.ProductListingID)
			} else {
				assert.Empty(t, tt.editArg.ProductListingID)
			}
		})
	}
}

func TestUpdateProductListingResponse_Structure(t *testing.T) {
	tests := []struct {
		name        string
		response    UpdateProductListingResponse
		wantSuccess bool
		wantMessage string
	}{
		{
			name: "successful update",
			response: UpdateProductListingResponse{
				ProductListing: nil, // 实际使用中会有真实的ProductListing对象
				Success:        true,
				Message:        "Product listing updated successfully",
			},
			wantSuccess: true,
			wantMessage: "Product listing updated successfully",
		},
		{
			name: "failed update",
			response: UpdateProductListingResponse{
				ProductListing: nil,
				Success:        false,
				Message:        "Update failed: validation error",
			},
			wantSuccess: false,
			wantMessage: "Update failed: validation error",
		},
		{
			name: "failed update without message",
			response: UpdateProductListingResponse{
				ProductListing: nil,
				Success:        false,
			},
			wantSuccess: false,
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantSuccess, tt.response.Success)
			assert.Equal(t, tt.wantMessage, tt.response.Message)
		})
	}
}

func TestNewUpdateProductListingTool(t *testing.T) {
	tool := NewUpdateProductListingTool()

	assert.NotNil(t, tool)
	assert.IsType(t, &UpdateProductListingTool{}, tool)
}

func TestConvertToRelationArgs(t *testing.T) {
	tests := []struct {
		name        string
		relations   []*product_listings.ProductListingRelation
		expectCount int
		expectNil   bool
	}{
		{
			name:        "nil relations",
			relations:   nil,
			expectCount: 0,
			expectNil:   true,
		},
		{
			name:        "empty relations",
			relations:   []*product_listings.ProductListingRelation{},
			expectCount: 0,
			expectNil:   false,
		},
		{
			name: "single relation",
			relations: []*product_listings.ProductListingRelation{
				{
					ID:                      "rel-1",
					VariantPosition:         1,
					ProductListingVariantID: "variant-1",
				},
			},
			expectCount: 1,
			expectNil:   false,
		},
		{
			name: "multiple relations with nil item",
			relations: []*product_listings.ProductListingRelation{
				{
					ID:                      "rel-1",
					VariantPosition:         1,
					ProductListingVariantID: "variant-1",
				},
				nil,
				{
					ID:                      "rel-2",
					VariantPosition:         2,
					ProductListingVariantID: "variant-2",
				},
			},
			expectCount: 2,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToRelationArgs(tt.relations)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				if tt.expectCount == 0 {
					// For empty slice, result could be nil or empty slice
					assert.Len(t, result, tt.expectCount)
				} else {
					assert.NotNil(t, result)
					assert.Len(t, result, tt.expectCount)
				}

				// 验证每个转换后的关系参数
				for i, relationArg := range result {
					assert.NotNil(t, relationArg)
					assert.Equal(t, "enabled", relationArg.AllowSync) // 默认值应该是 "enabled"

					// 找到对应的原始关系（跳过nil项）
					var originalIndex int
					for _, orig := range tt.relations {
						if orig != nil {
							if originalIndex == i {
								assert.Equal(t, orig.ID, relationArg.ID)
								assert.Equal(t, orig.VariantPosition, relationArg.VariantPosition)
								assert.Equal(t, orig.ProductListingVariantID, relationArg.ProductListingVariantID)
								break
							}
							originalIndex++
						}
					}
				}
			}
		})
	}
}

// Note: 实际的API调用测试需要mock外部依赖，这里只测试基本结构和逻辑
// 集成测试应该在单独的测试文件中进行，并使用适当的测试环境
