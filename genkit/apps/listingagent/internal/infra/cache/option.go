package cache

import (
	"github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/infra/redis"
)

type Option struct {
	RedisBusiness *redis.Options `json:"redis_business"  mapstructure:"redis_business" validate:"required"`
	RedisInfra    *redis.Options `json:"redis_infra"  mapstructure:"redis_infra" validate:"required"`
	//Local         *local.Option  `json:"local" mapstructure:"local" validate:"omitempty,required"`
}

func NewOption() *Option {
	return &Option{
		RedisBusiness: redis.NewRedisOption(),
		RedisInfra:    redis.NewRedisOption(),
		//Local:         local.NewOption(),
	}
}
