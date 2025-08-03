package lock

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"strings"
	"time"
)

const gLockPrefix = "lock"

var gRedisLock *redsync.Redsync

func InitRedisLock(client *redis.Client) error {
	pool := goredis.NewPool(client)
	gRedisLock = redsync.New(pool)
	return nil
}

func NewRedisLockMutex(expiry time.Duration, delay time.Duration, tries int,
	namespace string, subKeys ...string) *redsync.Mutex {
	key := strings.Join(append([]string{gLockPrefix, namespace}, subKeys...), "-")
	minDelay := delay.Milliseconds() / 2
	delayFunc := func(tries int) time.Duration {
		return time.Duration(rand.Intn(int(delay.Milliseconds()-minDelay))+int(minDelay)) * time.Millisecond
	}
	mutex := GetRedisLockClient().NewMutex(key, redsync.WithExpiry(expiry),
		redsync.WithRetryDelayFunc(delayFunc), redsync.WithTries(tries))
	return mutex
}

func NewRedisLockMutexByDefault(namespace string, subKeys ...string) *redsync.Mutex {
	return NewRedisLockMutex(30*time.Minute, 2*time.Second, 3, namespace, subKeys...)
}

// Lock Example
// 默认锁定 30 分钟，重试 3 次，每次重试间隔 2 秒
// mutex, err := lock.Lock("users", "xxxx")
//
//	if err != nil {
//		return []models.User{}, err
//	}
//
//	defer func(mutex *redsync.Mutex) {
//		_, err = mutex.Unlock()
//		if err != nil {
//			log.L(ctx).Warnw("unlock mutex failed.", err)
//		}
//	}(mutex)
func Lock(namespace string, subKeys ...string) (*redsync.Mutex, error) {
	mutex := NewRedisLockMutexByDefault(namespace, subKeys...)
	err := mutex.Lock()
	if err != nil {
		return nil, err
	}
	return mutex, nil
}

func GetRedisLockClient() *redsync.Redsync {
	return gRedisLock
}
