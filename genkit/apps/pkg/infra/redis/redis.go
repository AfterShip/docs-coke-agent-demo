package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(options *Options) (*redis.Client, error) {
	poolSize := options.PoolSize
	if options.PoolSize == 0 {
		poolSize = 120
	}

	cli := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    options.MasterName,
		SentinelAddrs: options.SentinelAddr,
		Password:      options.Password,
		DB:            options.DBNumber,
		PoolSize:      poolSize,
	})
	pong := cli.Ping(context.Background())
	if pong.Err() != nil {
		return nil, pong.Err()
	}
	return cli, nil
}
