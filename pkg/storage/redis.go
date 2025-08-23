package storage

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     100,
		MinIdleConns: 100,
		MaxRetries:   3,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolTimeout:  2 * time.Second,
	})
	return rdb
}
