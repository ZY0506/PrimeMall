package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DB struct {
		DSN          string `json:"DSN"`
		MaxIdleConns int    `json:"MaxIdleConns"`
		MaxOpenConns int    `json:"MaxOpenConns"`
	}
	Elasticsearch struct {
		Addresses []string `json:"Addresses"` // ES集群地址
		Username  string   `json:"Username"`
		Password  string   `json:"Password"`
	}
}
