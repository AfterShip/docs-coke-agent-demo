package externalapi

import (
	"context"
	"github.com/AfterShip/connectors-library/sdks/product_listings"
	"github.com/AfterShip/connectors-library/sdks/products_center"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi/aftership"
)

var listingClient *product_listings.Client
var productCenterClient *products_center.Client

func InitExternalAPIs(ctx context.Context, option Option) error {
	// Initialize AfterShip API
	var err error
	if listingClient, err = aftership.InitProductListingClient(ctx, option.AfterShipAPI); err != nil {
		return err
	}

	if productCenterClient, err = aftership.InitProductCenterClient(ctx, option.AfterShipAPI); err != nil {
		return err
	}
	// Initialize other external APIs here if needed
	return nil
}

func GetProductListingClient() *product_listings.Client {
	if listingClient == nil {
		panic("Product Listing client is not initialized. Call InitExternalAPIs first.")
	}
	return listingClient
}

func GetProductCenterClient() *products_center.Client {
	if productCenterClient == nil {
		panic("Product Center client is not initialized. Call InitExternalAPIs first.")
	}
	return productCenterClient
}
