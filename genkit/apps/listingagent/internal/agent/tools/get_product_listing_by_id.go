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

// GetProductListingByIDRequest represents the input for get_product_listing_by_id tool
type GetProductListingByIDRequest struct {
	ProductListingID string `json:"id" jsonschema:"required,description=The unique identifier of the product listing to retrieve"`
}

// GetProductListingByIDResponse represents the output of get_product_listing_by_id tool
type GetProductListingByIDResponse struct {
	ProductListing *product_listings.ProductListing `json:"product_listing"`
	Found          bool                             `json:"found"`
}

// GetProductListingByIDTool implements the get_product_listing_by_id tool
type GetProductListingByIDTool struct{}

// NewGetProductListingByIDTool creates a new get product listing by ID tool instance
func NewGetProductListingByIDTool() *GetProductListingByIDTool {
	return &GetProductListingByIDTool{}
}

// Name returns the tool name
func (g *GetProductListingByIDTool) Name() string {
	return "get_product_listing_by_id"
}

// Description returns the tool description
func (g *GetProductListingByIDTool) Description() string {
	return "Retrieve a specific product listing by its unique ID. Used to get detailed information about a single product listing for view, edit, or update operations."
}

// Define creates and registers the get product listing by ID tool with the given genkit client
func (g *GetProductListingByIDTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(client, g.Name(), g.Description(),
		func(toolCtx *ai.ToolContext, input GetProductListingByIDRequest) (string, error) {
			log.L(ctx).Info("get_product_listing_by_id tool called",
				zap.String("product_listing_id", input.ProductListingID))

			// 调用外部API获取产品列表详情
			productListing, err := externalapi.GetProductListingClient().ProductListing.
				GetByID(context.Background(), input.ProductListingID)
			if err != nil {
				log.L(ctx).Error("Failed to get product listing by ID",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))
				return "", err
			}

			// 构建响应
			response := GetProductListingByIDResponse{
				ProductListing: productListing,
				Found:          productListing != nil,
			}

			log.L(ctx).Info("get_product_listing_by_id tool response",
				zap.String("product_listing_id", input.ProductListingID),
				zap.Bool("found", response.Found))

			bytes, err := json.Marshal(response)
			if err != nil {
				log.L(ctx).Error("Failed to marshal product listing response",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))
				return "", err
			}
			return string(bytes), nil
		},
	)
}
