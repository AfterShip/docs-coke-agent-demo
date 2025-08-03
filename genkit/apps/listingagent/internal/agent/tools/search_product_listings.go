package tools

import (
	"context"
	"encoding/json"
	"github.com/AfterShip/connectors-library/sdks/product_listings"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi"

	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"go.uber.org/zap"
)

// GetProductListingsRequest represents the input for get_product_listings tool
type GetProductListingsRequest struct {
	Query string `json:"query" jsonschema:"required,description=The search query to find product listings (e.g., product title, keywords)"`
}

// GetProductListingsResponse represents the output of get_product_listings tool
type GetProductListingsResponse struct {
	ProductListings []ProductListingInfo `json:"product_listings"`
	Total           int                  `json:"total"`
}

// ProductListingInfo represents basic product listing information
type ProductListingInfo struct {
	ProductListing product_listings.ProductListing `json:"product_listing"`
}

// GetProductListingsTool implements the search_product_listings tool
type GetProductListingsTool struct{}

// NewGetProductListingsTool creates a new get product listings tool instance
func NewGetProductListingsTool() *GetProductListingsTool {
	return &GetProductListingsTool{}
}

// Name returns the tool name
func (g *GetProductListingsTool) Name() string {
	return "search_product_listings"
}

// Description returns the tool description
func (g *GetProductListingsTool) Description() string {
	return "Search and retrieve product listings from the system using natural language queries. This tool serves as the primary resource resolution mechanism for converting merchant-provided product identifiers (titles, keywords, SKUs, or partial names) into actionable product records with system IDs.\n\nThe tool accepts a single query parameter containing the search terms provided by merchants. It performs intelligent matching across product titles, descriptions, SKUs, and other searchable fields to locate relevant products within the merchant's catalog. The search algorithm prioritizes exact matches first, then performs fuzzy matching on keywords to ensure comprehensive coverage of potential matches.\n\nReturns structured product information including system IDs, titles, SKUs, current synchronization status, platform associations (Shopify/TikTok), and other essential metadata required for subsequent operations. Results are ranked by relevance and limited to prevent overwhelming the user interface.\n\nThis tool is essential for the three-stage workflow architecture, bridging the gap between natural language merchant requests and precise API operations. It supports all core business operations including product publishing, editing, activation, deactivation, and status queries by providing the accurate product identification required by downstream tools.\n\nThe tool handles various search scenarios including exact product name matching, partial title searches, SKU-based lookups, and keyword-driven discovery. It maintains context awareness to support batch operations and provides sufficient detail to enable confident product selection during the confirmation process.\n\nInput Parameter:\n- query (string): Search terms provided by the merchant, supporting product titles, keywords, SKUs, or any combination thereof"
}

// Define creates and registers the get product listings tool with the given genkit client
func (g *GetProductListingsTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool[GetProductListingsRequest, string](client, g.Name(), g.Description(),
		func(toolCtx *ai.ToolContext, input GetProductListingsRequest) (string, error) {
			log.L(ctx).Info("search_product_listings tool called",
				zap.String("query", input.Query))

			// 创建搜索参数
			params := product_listings.SearchProductListingParams{
				Query:          input.Query,
				OrganizationID: "45f9e1a1a77949d4ac9332bf9765ef7e",
				Page:           1,
			}

			// 调用外部API搜索产品列表
			resp, err := externalapi.GetProductListingClient().ProductListing.
				Search(context.Background(), &params)
			if err != nil {
				log.L(ctx).Error("Failed to search product listings",
					zap.String("query", input.Query),
					zap.Error(err))
				return "", err
			}

			// 转换响应为工具输出格式
			var productListings []ProductListingInfo
			if resp != nil && resp.ProductListings != nil {
				for _, listing := range resp.ProductListings {
					productInfo := ProductListingInfo{
						ProductListing: listing,
					}
					productListings = append(productListings, productInfo)
				}
			}

			response := GetProductListingsResponse{
				ProductListings: productListings,
				Total:           len(productListings),
			}

			log.L(ctx).Info("search_product_listings tool response",
				zap.String("query", input.Query),
				zap.Int("total_found", response.Total))

			bytes, err := json.Marshal(response)
			if err != nil {
				log.L(ctx).Error("Failed to marshal product listings response",
					zap.String("query", input.Query),
					zap.Error(err))
				return "", err
			}
			return string(bytes), nil
		},
	)
}
