package aftership

import (
	"context"
	"github.com/AfterShip/connectors-library/httpx"
	"github.com/AfterShip/connectors-library/sdks/product_listings"
	"github.com/AfterShip/connectors-library/sdks/products_center"
	"time"
)

func InitProductListingClient(ctx context.Context, option *Option) (*product_listings.Client, error) {
	return product_listings.NewClient(httpx.NewClient(option.ProductListingUrl, option.APIKey, httpx.WithTimeout(10*time.Minute))), nil
}

func InitProductCenterClient(ctx context.Context, option *Option) (*products_center.Client, error) {
	return products_center.NewClient(httpx.NewClient(option.ProductCenterUrl, option.APIKey, httpx.WithTimeout(10*time.Minute))), nil
}
