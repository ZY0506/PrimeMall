package rdb

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// InitRedis 初始化redis
func InitRedis(addr string, password string, db, poolSize, minIdleConns int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}
