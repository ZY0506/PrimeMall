package svc

import (
	"fmt"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	Client          *redis.Client
	CouponModel     model.CouponModel
	UserCouponModel model.UserCouponModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	fmt.Println("初始化数据库连接：", c.DB.DSN)
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("Redis初始化失败，错误：%v", err.Error())
		panic(err)
	}
	return &ServiceContext{
		Config:          c,
		DB:              db,
		Client:          client,
		CouponModel:     model.NewCouponModel(db),
		UserCouponModel: model.NewUserCouponModel(db),
	}
}
