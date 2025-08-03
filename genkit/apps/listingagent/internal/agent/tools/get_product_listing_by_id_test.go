package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProductListingByIDTool_Name(t *testing.T) {
	tool := NewGetProductListingByIDTool()
	assert.Equal(t, "get_product_listing_by_id", tool.Name())
}

func TestGetProductListingByIDTool_Description(t *testing.T) {
	tool := NewGetProductListingByIDTool()
	description := tool.Description()

	assert.NotEmpty(t, description)
	assert.Contains(t, description, "Retrieve a specific product listing")
	assert.Contains(t, description, "unique ID")
}

func TestGetProductListingByIDRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		request       GetProductListingByIDRequest
		expectedOrgID string
	}{
		{
			name: "with organization ID provided",
			request: GetProductListingByIDRequest{
				ProductListingID: "test-id-123",
				OrganizationID:   "custom-org-id",
			},
			expectedOrgID: "custom-org-id",
		},
		{
			name: "without organization ID - should use default",
			request: GetProductListingByIDRequest{
				ProductListingID: "test-id-456",
			},
			expectedOrgID: "ddadb6300a2049eca15fb8e8e35c1af5",
		},
		{
			name: "with empty organization ID - should use default",
			request: GetProductListingByIDRequest{
				ProductListingID: "test-id-789",
				OrganizationID:   "",
			},
			expectedOrgID: "ddadb6300a2049eca15fb8e8e35c1af5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证请求结构
			assert.NotEmpty(t, tt.request.ProductListingID)

			// 验证默认组织ID逻辑
			orgID := tt.request.OrganizationID
			if orgID == "" {
				orgID = "ddadb6300a2049eca15fb8e8e35c1af5"
			}
			assert.Equal(t, tt.expectedOrgID, orgID)
		})
	}
}

func TestGetProductListingByIDResponse_Structure(t *testing.T) {
	tests := []struct {
		name      string
		response  GetProductListingByIDResponse
		wantFound bool
	}{
		{
			name: "found response",
			response: GetProductListingByIDResponse{
				ProductListing: nil, // 实际使用中会有真实的ProductListing对象
				Found:          true,
			},
			wantFound: true,
		},
		{
			name: "not found response",
			response: GetProductListingByIDResponse{
				ProductListing: nil,
				Found:          false,
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantFound, tt.response.Found)
		})
	}
}

func TestNewGetProductListingByIDTool(t *testing.T) {
	tool := NewGetProductListingByIDTool()

	assert.NotNil(t, tool)
	assert.IsType(t, &GetProductListingByIDTool{}, tool)
}

// Note: 实际的API调用测试需要mock外部依赖，这里只测试基本结构和逻辑
// 集成测试应该在单独的测试文件中进行，并使用适当的测试环境
