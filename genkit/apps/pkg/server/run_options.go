package server

import (
	"github.com/gin-gonic/gin"
)

// RunOptions contains the options while running a generic api server.
type RunOptions struct {
	Mode        string   `json:"mode"        mapstructure:"mode"`
	DefaultAPIs []string `json:"default_api"  mapstructure:"default_apis"`
	Middlewares []string `json:"middlewares" mapstructure:"middlewares"`
	// default: true
	EnableProfiling bool `json:"profiling"      mapstructure:"profiling"`
	// default: true
	EnableMetrics bool `json:"metrics" mapstructure:"metrics"`
}

// NewRunOptions creates a new RunOptions object with default parameters.
func NewRunOptions() *RunOptions {
	return &RunOptions{
		Mode:        gin.ReleaseMode,
		Middlewares: []string{},
		DefaultAPIs: []string{
			"version",
			"healthz",
			"whoami",
		},
		EnableMetrics:   true,
		EnableProfiling: true,
	}
}
