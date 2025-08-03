package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	storeRedis "github.com/ulule/limiter/v3/drivers/store/redis"
)

const (
	LimiterPrefixTemplate = "limiter:%s"
)

func LimitWithMemory(rateFromFormatted string,
	limitReachedHandler mgin.LimitReachedHandler,
	errorHandler mgin.ErrorHandler) gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted(rateFromFormatted)
	if err != nil {
		panic(err)
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)
	errorHandleOption := mgin.WithErrorHandler(errorHandler)
	limitReachedHandlerOption := mgin.WithLimitReachedHandler(limitReachedHandler)

	middleware := mgin.NewMiddleware(instance, limitReachedHandlerOption, errorHandleOption)
	return middleware
}

func LimitWithRedis(rateFromFormatted string,
	client *redis.Client,
	prefix string,
	limitReachedHandler mgin.LimitReachedHandler,
	errorHandler mgin.ErrorHandler) gin.HandlerFunc {

	rate, err := limiter.NewRateFromFormatted(rateFromFormatted)
	if err != nil {
		panic(err)
	}

	store, err := storeRedis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: fmt.Sprintf(LimiterPrefixTemplate, prefix),
	})

	instance := limiter.New(store, rate)
	errorHandleOption := mgin.WithErrorHandler(errorHandler)
	limitReachedHandlerOption := mgin.WithLimitReachedHandler(limitReachedHandler)
	excludeKeyOption := mgin.WithExcludedKey(func(key string) bool {
		return false
	})

	middleware := mgin.NewMiddleware(instance, limitReachedHandlerOption, errorHandleOption, excludeKeyOption)
	return middleware
}
