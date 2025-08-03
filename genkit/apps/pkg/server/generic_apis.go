package server

import (
	version "github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/project-version"
	"github.com/gin-gonic/gin"
	"github.com/marmotedu/component-base/pkg/core"
)

var DefaultAPIs = defaultAPIs()

type WithAPI func(s *GenericAPIServer)

func WithGetVersion() WithAPI {
	return func(s *GenericAPIServer) {
		s.GET("/version", func(c *gin.Context) {
			core.WriteResponse(c, nil, version.Get())
		})
	}
}

func WithGetWhoami() WithAPI {
	return func(s *GenericAPIServer) {
		s.GET("/whoami", func(c *gin.Context) {
			core.WriteResponse(c, nil, version.Get())
		})
	}
}

func WithGetHealthz() WithAPI {
	return func(s *GenericAPIServer) {
		s.GET("/healthz", func(c *gin.Context) {
			core.WriteResponse(c, nil, map[string]string{"status": "ok"})
		})
	}
}

func defaultAPIs() map[string]WithAPI {
	return map[string]WithAPI{
		"healthz": WithGetHealthz(),
		"version": WithGetVersion(),
		"whoami":  WithGetWhoami(),
	}
}
