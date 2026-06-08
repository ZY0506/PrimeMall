package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	JWT struct {
		Secret             string `json:"Secret"` // 使用 json 标签
		Issuer             string `json:"Issuer"` // 使用 json 标签
		AccessTokenExpire  int64  `json:"AccessTokenExpire"`
		RefreshTokenExpire int64  `json:"RefreshTokenExpire"`
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
	Snowflake struct {
		NodeID int64 `json:"NodeID"`
	}
}
