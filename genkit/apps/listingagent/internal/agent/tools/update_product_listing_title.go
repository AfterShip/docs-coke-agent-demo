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

// UpdateProductListingTitleRequest represents the input for update_product_listing_title tool
type UpdateProductListingTitleRequest struct {
	ProductListingID string `json:"id" jsonschema:"required,description=The unique identifier of the product listing to update"`
	Title            string `json:"title" jsonschema:"required,description=The new title for the product listing"`
}

// UpdateProductListingTitleResponse represents the output of update_product_listing_title tool
type UpdateProductListingTitleResponse struct {
	ProductListing *product_listings.ProductListing `json:"product_listing"`
	Success        bool                             `json:"success"`
	Message        string                           `json:"message,omitempty"`
	OldTitle       string                           `json:"old_title,omitempty"`
	NewTitle       string                           `json:"new_title,omitempty"`
}

// UpdateProductListingTitleTool implements the update_product_listing_title tool
type UpdateProductListingTitleTool struct{}

// NewUpdateProductListingTitleTool creates a new update product listing title tool instance
func NewUpdateProductListingTitleTool() *UpdateProductListingTitleTool {
	return &UpdateProductListingTitleTool{}
}

// Name returns the tool name
func (u *UpdateProductListingTitleTool) Name() string {
	return "update_product_listing_title"
}

// Description returns the tool description
func (u *UpdateProductListingTitleTool) Description() string {
	return "Update the title of an existing product listing. This tool specifically updates only the product title while preserving all other product information."
}

// Define creates and registers the update product listing title tool with the given genkit client
func (u *UpdateProductListingTitleTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(client, u.Name(), u.Description(),
		func(toolCtx *ai.ToolContext, input UpdateProductListingTitleRequest) (string, error) {
			log.L(ctx).Info("update_product_listing_title tool called",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("new_title", input.Title))

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
				response := UpdateProductListingTitleResponse{
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

				response := UpdateProductListingTitleResponse{
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

			// 记录旧标题
			oldTitle := existingListing.Product.Title

			// 创建更新后的Product，只修改Title字段
			updatedProduct := existingListing.Product
			updatedProduct.Title = input.Title

			// 创建更新参数，保留现有数据并仅更新Product的Title字段
			updateArg := &product_listings.UpdateArg{
				ID:                    input.ProductListingID,
				ProductsCenterProduct: existingListing.ProductsCenterProduct,
				Settings:              existingListing.Settings,
				Product:               updatedProduct,
				Relations:             convertToRelationArgs(existingListing.Relations),
				NeedPublish:           true, // 默认需要发布
			}

			log.L(ctx).Info("Updating product listing title",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("old_title", oldTitle),
				zap.String("new_title", input.Title))

			// 调用外部API更新产品列表
			updatedListing, err := externalapi.GetProductListingClient().ProductListing.
				Update(context.Background(), input.ProductListingID, updateArg)
			if err != nil {
				log.L(ctx).Error("Failed to update product listing title",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))

				// 返回失败响应
				response := UpdateProductListingTitleResponse{
					Success:  false,
					Message:  err.Error(),
					OldTitle: oldTitle,
					NewTitle: input.Title,
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
			response := UpdateProductListingTitleResponse{
				ProductListing: updatedListing,
				Success:        true,
				Message:        "Product listing title updated successfully",
				OldTitle:       oldTitle,
				NewTitle:       input.Title,
			}

			log.L(ctx).Info("update_product_listing_title tool response - success",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("old_title", oldTitle),
				zap.String("new_title", input.Title),
				zap.Bool("success", response.Success))

			bytes, err := json.Marshal(response)
			if err != nil {
				log.L(ctx).Error("Failed to marshal product listing title update response",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))
				return "", err
			}
			return string(bytes), nil
		},
	)
}
