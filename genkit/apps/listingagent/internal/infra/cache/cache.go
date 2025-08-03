package cache

import (
	"context"
	storeredis "github.com/AfterShip/docs-coke-agent-demo/genkit/apps/pkg/infra/redis"
	"github.com/AfterShip/docs-coke-agent-demo/genkit/pkg/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	gRedisBusiness *redis.Client
	gRedisInfra    *redis.Client
)

func initRedisClient(ctx context.Context, opt *Option) error {
	//业务使用的 redis 实例
	var err error
	gRedisBusiness, err = storeredis.NewRedisClient(opt.RedisBusiness)
	if err != nil {
		return err
	}

	//Limiter、Lock 等使用的 redis 实例
	gRedisInfra, err = storeredis.NewRedisClient(opt.RedisInfra)
	if err != nil {
		return err
	}

	return nil
}

// GetRedisBusiness returns the redis client for business.
func GetRedisBusiness() *redis.Client {
	return gRedisBusiness
}

// GetRedisInfra returns the redis client for common, such as limiter, lock.
func GetRedisInfra() *redis.Client {
	return gRedisInfra
}

func InitCache(ctx context.Context, opt *Option) error {
	return initRedisClient(ctx, opt)
}

func Close() {
	var err error
	if GetRedisBusiness() != nil {
		if err = GetRedisBusiness().Close(); err != nil {
			log.Warn("redis business close error: %v", zap.Error(err))
		}
	}

	if GetRedisInfra() != nil {
		if err = GetRedisInfra().Close(); err != nil {
			log.Warn("redis common close error: %v", zap.Error(err))
		}
	}
}
