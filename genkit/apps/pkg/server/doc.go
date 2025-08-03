package server // import "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"

// ErrorResponse The response means there is an error.
// swagger:response ErrorResponse
type ErrorResponse struct {
	// in:body
	Body struct {
		// Meta summary information
		//
		// Required: true
		Meta Meta `json:"meta"`
	}
}
