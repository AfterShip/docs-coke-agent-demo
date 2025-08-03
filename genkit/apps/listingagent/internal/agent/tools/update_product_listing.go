package tools

import (
	"context"
	"encoding/json"
	"github.com/AfterShip/connectors-library/sdks/product_listings"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi"
	agentmodel "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/model/agent"

	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"go.uber.org/zap"
)

// UpdateProductListingRequest represents the input for update_product_listing tool
type UpdateProductListingRequest struct {
	ProductListingID string                          `json:"product_listing_id" jsonschema:"required,description=The unique identifier of the product listing to update"`
	ProductListing   product_listings.ProductListing `json:"product_listing" jsonschema:"required,description=The updated product listing data"`
	OrganizationID   string                          `json:"organization_id,omitempty" jsonschema:"description=The organization ID (defaults to ddadb6300a2049eca15fb8e8e35c1af5 if not provided)"`
}

// UpdateProductListingResponse represents the output of update_product_listing tool
type UpdateProductListingResponse struct {
	ProductListing *product_listings.ProductListing `json:"product_listing"`
	Success        bool                             `json:"success"`
	Message        string                           `json:"message,omitempty"`
}

// UpdateProductListingTool implements the update_product_listing tool
type UpdateProductListingTool struct{}

// NewUpdateProductListingTool creates a new update product listing tool instance
func NewUpdateProductListingTool() *UpdateProductListingTool {
	return &UpdateProductListingTool{}
}

// Name returns the tool name
func (u *UpdateProductListingTool) Name() string {
	return "update_product_listing"
}

// Description returns the tool description
func (u *UpdateProductListingTool) Description() string {
	return "Update an existing product listing with new information. Used to modify product details, status, pricing, or other attributes of a specific product listing."
}

// Define creates and registers the update product listing tool with the given genkit client
func (u *UpdateProductListingTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(client, u.Name(), u.Description(),
		func(toolCtx *ai.ToolContext, input agentmodel.EditArg) (string, error) {
			log.L(ctx).Info("update_product_listing tool called",
				zap.String("product_listing_id", input.ProductListingID))

			// 首先获取现有的产品列表数据
			log.L(ctx).Info("Getting existing product listing data",
				zap.String("product_listing_id", input.ProductListingID))

			existingListing, err := externalapi.GetProductListingClient().ProductListing.
				GetByID(context.Background(), input.ProductListingID)
			if err != nil {
				log.L(ctx).Error("Failed to get existing product listing",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))

				// 返回失败响应
				response := UpdateProductListingResponse{
					Success: false,
					Message: "Failed to get existing product listing: " + err.Error(),
				}

				bytes, marshalErr := json.Marshal(response)
				if marshalErr != nil {
					log.L(ctx).Error("Failed to marshal error response",
						zap.String("product_listing_id", input.ProductListingID),
						zap.Error(marshalErr))
					return "", marshalErr
				}
				return string(bytes), nil
			}

			if existingListing == nil {
				log.L(ctx).Error("Product listing not found",
					zap.String("product_listing_id", input.ProductListingID))

				response := UpdateProductListingResponse{
					Success: false,
					Message: "Product listing not found",
				}

				bytes, marshalErr := json.Marshal(response)
				if marshalErr != nil {
					log.L(ctx).Error("Failed to marshal not found response",
						zap.String("product_listing_id", input.ProductListingID),
						zap.Error(marshalErr))
					return "", marshalErr
				}
				return string(bytes), nil
			}

			// 创建更新参数，保留现有数据并仅更新Product字段
			updateArg := &product_listings.UpdateArg{
				ID:                    input.ProductListingID,
				ProductsCenterProduct: existingListing.ProductsCenterProduct,
				Settings:              existingListing.Settings,
				Product:               input.Product, // 使用EditArg中的Product更新
				Relations:             convertToRelationArgs(existingListing.Relations),
				NeedPublish:           true, // 默认需要发布
			}

			log.L(ctx).Info("Updating product listing with merged data",
				zap.String("product_listing_id", input.ProductListingID))

			// 调用外部API更新产品列表
			updatedListing, err := externalapi.GetProductListingClient().ProductListing.
				Update(context.Background(), input.ProductListingID, updateArg)
			if err != nil {
				log.L(ctx).Error("Failed to update product listing",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))

				// 返回失败响应
				response := UpdateProductListingResponse{
					Success: false,
					Message: err.Error(),
				}

				bytes, marshalErr := json.Marshal(response)
				if marshalErr != nil {
					log.L(ctx).Error("Failed to marshal error response",
						zap.String("product_listing_id", input.ProductListingID),
						zap.Error(marshalErr))
					return "", marshalErr
				}
				return string(bytes), nil
			}

			// 构建成功响应
			response := UpdateProductListingResponse{
				ProductListing: updatedListing,
				Success:        true,
				Message:        "Product listing updated successfully",
			}

			log.L(ctx).Info("update_product_listing tool response - success",
				zap.String("product_listing_id", input.ProductListingID),
				zap.Bool("success", response.Success))

			bytes, err := json.Marshal(response)
			if err != nil {
				log.L(ctx).Error("Failed to marshal product listing update response",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))
				return "", err
			}
			return string(bytes), nil
		},
	)
}

// convertToRelationArgs converts ProductListingRelation slice to ProductListingRelationArg slice
func convertToRelationArgs(relations []*product_listings.ProductListingRelation) []*product_listings.ProductListingRelationArg {
	if relations == nil {
		return nil
	}

	var relationArgs []*product_listings.ProductListingRelationArg
	for _, relation := range relations {
		if relation != nil {
			relationArg := &product_listings.ProductListingRelationArg{
				ID:                      relation.ID,
				VariantPosition:         relation.VariantPosition,
				ProductListingVariantID: relation.ProductListingVariantID,
				SalesChannelVariant:     relation.SalesChannelVariant,
				ProductsCenterVariant:   relation.ProductsCenterVariant,
				AllowSync:               "enabled", // 默认启用同步
			}
			relationArgs = append(relationArgs, relationArg)
		}
	}
	return relationArgs
}
