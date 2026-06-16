package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	ProductRpc   zrpc.RpcClientConf
	UserRpc      zrpc.RpcClientConf
	MarketingRpc zrpc.RpcClientConf
	PaymentRpc   zrpc.RpcClientConf
	DB           struct {
		DSN          string `json:"DSN"`
		MaxIdleConns int    `json:"MaxIdleConns"`
		MaxOpenConns int    `json:"MaxOpenConns"`
	}
	RDB struct {
		RedisAddr         string `json:"Addr"`
		RedisPassword     string `json:"RedisPassword"`
		RedisDB           int    `json:"RedisDB"`
		RedisPoolSize     int    `json:"RedisPoolSize"`
		RedisMinIdleConns int    `json:"RedisMinIdleConns"`
	}
	Snowflake struct {
		NodeID int64 `json:"NodeID"`
	}
	RabbitMQ struct {
		URL string `json:"URL"`
	}
}
