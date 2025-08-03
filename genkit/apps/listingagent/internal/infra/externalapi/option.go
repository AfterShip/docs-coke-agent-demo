package externalapi

import "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/listingagent/internal/infra/externalapi/aftership"

type Option struct {
	AfterShipAPI *aftership.Option `json:"aftership" mapstructure:"aftership" validate:"required"`
}

func NewOption() *Option {
	return &Option{
		AfterShipAPI: aftership.NewOption(),
	}
}
