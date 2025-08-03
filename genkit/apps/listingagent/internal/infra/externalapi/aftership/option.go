package aftership

type Option struct {
	ProductListingUrl string `json:"product_listing_url" mapstructure:"product_listing_url" validate:"required"`
	ProductCenterUrl  string `json:"product_center_url" mapstructure:"product_center_url" validate:"required"`
	APIKey            string //从 env 中读取，不在 config 文件配置；
}

func NewOption() *Option {
	return &Option{
		ProductListingUrl: "",
		ProductCenterUrl:  "",
		APIKey:            "",
	}
}
