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

// UpdateProductListingDescriptionRequest represents the input for update_product_listing_description tool
type UpdateProductListingDescriptionRequest struct {
	ProductListingID string `json:"id" jsonschema:"required,description=The unique identifier of the product listing to update"`
	Description      string `json:"description" jsonschema:"required,description=The new description for the product listing"`
}

// UpdateProductListingDescriptionResponse represents the output of update_product_listing_description tool
type UpdateProductListingDescriptionResponse struct {
	ProductListing *product_listings.ProductListing `json:"product_listing"`
	Success        bool                             `json:"success"`
	Message        string                           `json:"message,omitempty"`
	OldDescription string                           `json:"old_description,omitempty"`
	NewDescription string                           `json:"new_description,omitempty"`
}

// UpdateProductListingDescriptionTool implements the update_product_listing_description tool
type UpdateProductListingDescriptionTool struct{}

// NewUpdateProductListingDescriptionTool creates a new update product listing description tool instance
func NewUpdateProductListingDescriptionTool() *UpdateProductListingDescriptionTool {
	return &UpdateProductListingDescriptionTool{}
}

// Name returns the tool name
func (u *UpdateProductListingDescriptionTool) Name() string {
	return "update_product_listing_description"
}

// Description returns the tool description
func (u *UpdateProductListingDescriptionTool) Description() string {
	return "Update the description of an existing product listing. This tool specifically updates only the product description while preserving all other product information."
}

// Define creates and registers the update product listing description tool with the given genkit client
func (u *UpdateProductListingDescriptionTool) Define(ctx context.Context, client *genkit.Genkit) ai.ToolRef {
	return genkit.DefineTool(client, u.Name(), u.Description(),
		func(toolCtx *ai.ToolContext, input UpdateProductListingDescriptionRequest) (string, error) {
			log.L(ctx).Info("update_product_listing_description tool called",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("new_description", input.Description))

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
				response := UpdateProductListingDescriptionResponse{
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

				response := UpdateProductListingDescriptionResponse{
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

			// 记录旧描述
			oldDescription := existingListing.Product.Description

			// 创建更新后的Product，只修改Description字段
			updatedProduct := existingListing.Product
			updatedProduct.Description = input.Description

			// 创建更新参数，保留现有数据并仅更新Product的Description字段
			updateArg := &product_listings.UpdateArg{
				ID:                    input.ProductListingID,
				ProductsCenterProduct: existingListing.ProductsCenterProduct,
				Settings:              existingListing.Settings,
				Product:               updatedProduct,
				Relations:             convertToRelationArgs(existingListing.Relations),
				NeedPublish:           true, // 默认需要发布
			}

			log.L(ctx).Info("Updating product listing description",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("old_description", oldDescription),
				zap.String("new_description", input.Description))

			// 调用外部API更新产品列表
			updatedListing, err := externalapi.GetProductListingClient().ProductListing.
				Update(context.Background(), input.ProductListingID, updateArg)
			if err != nil {
				log.L(ctx).Error("Failed to update product listing description",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))

				// 返回失败响应
				response := UpdateProductListingDescriptionResponse{
					Success:        false,
					Message:        err.Error(),
					OldDescription: oldDescription,
					NewDescription: input.Description,
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
			response := UpdateProductListingDescriptionResponse{
				ProductListing: updatedListing,
				Success:        true,
				Message:        "Product listing description updated successfully",
				OldDescription: oldDescription,
				NewDescription: input.Description,
			}

			log.L(ctx).Info("update_product_listing_description tool response - success",
				zap.String("product_listing_id", input.ProductListingID),
				zap.String("old_description", oldDescription),
				zap.String("new_description", input.Description),
				zap.Bool("success", response.Success))

			bytes, err := json.Marshal(response)
			if err != nil {
				log.L(ctx).Error("Failed to marshal product listing description update response",
					zap.String("product_listing_id", input.ProductListingID),
					zap.Error(err))
				return "", err
			}
			return string(bytes), nil
		},
	)
}
