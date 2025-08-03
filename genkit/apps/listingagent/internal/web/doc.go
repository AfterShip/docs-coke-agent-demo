package web

// Package docs API V1.
//
// Identity and Access Management System.
//
//     Schemes: http, https
//     Host: iam.api.xxxx.com
//     BasePath: /v1
//     Version: 1.0.0
//	   License: MIT https://opensource.org/licenses/MIT
//	   Contact: xxxx <xxx@gmail.com> http://xxxxx.com
//
//     User:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - basic
//     - api_key
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//    api_key:
//      type: apiKey
//      name: Authorization
//      in: header
//
// swagger:meta

import "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/server"

// ErrorResponse The response means there is an error.
// swagger:response ErrorResponse
type ErrorResponse struct {
	// in:body
	Body struct {
		// Meta summary information
		//
		// Required: true
		Meta server.Meta `json:"meta"`
	}
}
