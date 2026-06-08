package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	JWT struct {
		Secret            string `json:"Secret"`
		Issuer            string `json:"Issuer"`
		AccessTokenExpire int64  `json:"AccessTokenExpire"`
	}
	DB struct {
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
	UserRpc    zrpc.RpcClientConf `json:"UserRpc"`
	ProductRpc zrpc.RpcClientConf `json:"ProductRpc"`
	OrderRpc   zrpc.RpcClientConf `json:"OrderRpc"`
}
