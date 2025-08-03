// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package server

import (
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// RecommendedHomeDir defines the default directory used to place all iam service configurations.
	RecommendedHomeDir = ".iam"

	// RecommendedEnvPrefix defines the ENV prefix used by all iam service.
	RecommendedEnvPrefix = "IAM"
)

// Config is a structure used to configure a GenericAPIServer.
// Its members are sorted roughly in order of importance for composers.
type Config struct {
	SecureServing   *SecureServingOptions
	InsecureServing *InsecureServingOptions
	Jwt             *JwtInfo
	RunInfo         *RunOptions
}

// JwtInfo defines jwt fields used to create jwt authentication middleware.
type JwtInfo struct {
	// defaults to "iam jwt"
	Realm string
	// defaults to empty
	Key string
	// defaults to one hour
	Timeout time.Duration
	// defaults to zero
	MaxRefresh time.Duration
}

// NewConfig returns a Config struct with the default values.
func NewConfig() *Config {
	return &Config{
		InsecureServing: NewInsecureServingOptions(),
		SecureServing:   NewSecureServingOptions(),
		RunInfo:         NewRunOptions(),
		Jwt: &JwtInfo{
			Realm:      "iam jwt",
			Timeout:    1 * time.Hour,
			MaxRefresh: 1 * time.Hour,
		},
	}
}

// CompletedConfig is the completed configuration for GenericAPIServer.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data and can be derived
// from other fields. If you're going to `ApplyOptions`, do that first. It's mutating the receiver.
func (c *Config) Complete() CompletedConfig {
	return CompletedConfig{c}
}

// New returns a new instance of GenericAPIServer from the given config.
func (c CompletedConfig) New() (*GenericAPIServer, error) {
	gin.SetMode(c.RunInfo.Mode)

	s := &GenericAPIServer{
		middlewares:         c.RunInfo.Middlewares,
		defaultAPIs:         c.RunInfo.DefaultAPIs,
		mode:                c.RunInfo.Mode,
		SecureServingInfo:   c.SecureServing,
		InsecureServingInfo: c.InsecureServing,
		enableMetrics:       c.RunInfo.EnableMetrics,
		enableProfiling:     c.RunInfo.EnableProfiling,
		Engine:              gin.New(),
	}

	//init the server.
	initGenericAPIServer(s)
	return s, nil
}
