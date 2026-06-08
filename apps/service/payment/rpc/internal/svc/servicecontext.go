package svc

import (
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/order"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/orderinternal"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/snowflakes"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                  config.Config
	DB                      sqlx.SqlConn
	Client                  *redis.Client
	IDGenerator             *snowflakes.Generator
	PaymentModel            model.PaymentModel
	PaymentCallbackLogModel model.PaymentCallbackLogModel
	RefundModel             model.RefundModel
	OrderRpc                order.Order
	OrderInternalRpc        orderinternal.OrderInternal
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := database.InitDB(c.DB.DSN, c.DB.MaxIdleConns, c.DB.MaxOpenConns)
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
		Config:                  c,
		DB:                      db,
		Client:                  client,
		IDGenerator:             IDGenerator,
		PaymentModel:            model.NewPaymentModel(db),
		PaymentCallbackLogModel: model.NewPaymentCallbackLogModel(db),
		RefundModel:             model.NewRefundModel(db),
		OrderRpc:                order.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		OrderInternalRpc:        orderinternal.NewOrderInternal(zrpc.MustNewClient(c.OrderRpc)),
	}
}
