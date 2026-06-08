// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserRpc      zrpc.RpcClientConf
	ProductRpc   zrpc.RpcClientConf
	OrderRpc     zrpc.RpcClientConf
	PaymentRpc   zrpc.RpcClientConf
	MarketingRpc zrpc.RpcClientConf
	SearchRpc    zrpc.RpcClientConf
	JwtAuth      struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshSecret string
		RefreshExpire int64
		Issuer        string
	}
	RDB struct {
		RedisAddr         string `json:"Addr"`
		RedisPassword     string `json:"RedisPassword"`
		RedisDB           int    `json:"RedisDB"`
		RedisPoolSize     int    `json:"RedisPoolSize"`
		RedisMinIdleConns int    `json:"RedisMinIdleConns"`
	}
	OSSConfig struct {
		AccessKeyId     string   `json:"AccessKeyId"`
		AccessKeySecret string   `json:"AccessKeySecret"`
		Endpoint        string   `json:"Endpoint"`
		BucketName      string   `json:"BucketName"`
		Host            string   `json:"Host"`
		UploadDir       string   `json:"UploadDir"`
		ExpireTime      int64    `json:"ExpireTime"`
		CallbackUrl     string   `json:"CallbackUrl"`
		MaxFileSize     int64    `json:"MaxFileSize"`
		AllowedExts     []string `json:"AllowedExts"`
	}
	CasbinConf struct {
		ModelPath  string
		PolicyPath string
	}
	RateLimit struct {
		PerIPSpeed int
		PerIPBurst int
	}
}
