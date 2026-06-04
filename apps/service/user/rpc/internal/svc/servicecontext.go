package svc

import (
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/captcha"
	"github.com/ZY0506/PrimeMall/common/snowflakes"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	DB             sqlx.SqlConn
	Client         *redis.Client
	CaptchaSvc     *captcha.Service
	UserModel      model.UserModel
	AddressModel   model.UserAddressModel
	LoginLogModel  model.UserLoginLogModel
	PunishLogModel model.UserPunishLogModel
	IDGenerator    *snowflakes.Generator
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化db
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
	// 初始化redis
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("redis init failed,error:%v", err.Error())
		panic(err)
	}
	IDGenerator, err := snowflakes.NewGenerator(c.Snowflake.NodeID)
	if err != nil {
		logx.Errorf("snowflake init failed,error:%v", err.Error())
		panic(err)
	}
	return &ServiceContext{
		Config:         c,
		DB:             db,
		Client:         client,
		CaptchaSvc:     captcha.NewService(client),
		UserModel:      model.NewUserModel(db),
		AddressModel:   model.NewUserAddressModel(db),
		LoginLogModel:  model.NewUserLoginLogModel(db),
		PunishLogModel: model.NewUserPunishLogModel(db),
		IDGenerator:    IDGenerator,
	}
}
