package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	order2 "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productinternal"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/userinternal"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/snowflakes"
	"github.com/ZY0506/PrimeMall/pkg/database"
	"github.com/ZY0506/PrimeMall/pkg/mq/rabbitmq"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"time"
)

type ServiceContext struct {
	Config             config.Config
	DB                 sqlx.SqlConn
	Client             *redis.Client
	AfterSaleModel     model.AfterSaleModel
	AfterSaleItemModel model.AfterSaleItemModel
	CartModel          model.CartModel
	OrderInfoModel     model.OrderInfoModel
	OrderItemModel     model.OrderItemModel
	ProductRpc         productinternal.ProductInternal
	UserRpc            userinternal.UserInternal
	IDGenerator        *snowflakes.Generator
	MQClient           *rabbitmq.Client
	ctx                context.Context
	cancelFunc         context.CancelFunc
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
	mqClient, err := rabbitmq.NewClient(c.RabbitMQ.URL)
	if err != nil {
		logx.Errorf("rabbitmq init failed,error:%v", err.Error())
		panic(err)
	}

	err = mqClient.InitDelayQueue(constants.ORDER_DELAY_QUEUE, constants.ORDER_DLX_EXCHANGE, constants.ORDER_TIMEOUT_ROUTING_KEY, constants.ORDER_TTL_MS)
	if err != nil {
		logx.Errorf("rabbitmq init failed,error:%v", err.Error())
		panic(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	s := &ServiceContext{
		Config:             c,
		DB:                 db,
		Client:             client,
		AfterSaleModel:     model.NewAfterSaleModel(db),
		AfterSaleItemModel: model.NewAfterSaleItemModel(db),
		CartModel:          model.NewCartModel(db),
		OrderInfoModel:     model.NewOrderInfoModel(db),
		OrderItemModel:     model.NewOrderItemModel(db),
		ProductRpc:         productinternal.NewProductInternal(zrpc.MustNewClient(c.ProductRpc)),
		UserRpc:            userinternal.NewUserInternal(zrpc.MustNewClient(c.UserRpc)),
		IDGenerator:        IDGenerator,
		MQClient:           mqClient,
		ctx:                ctx,
		cancelFunc:         cancelFunc,
	}

	// 注册所有的消费者
	s.registerConsumers()

	return s
}

func (s *ServiceContext) Close() {
	s.cancelFunc()
	if s.MQClient != nil {
		_ = s.MQClient.Close()
	}
	return
}

func (s *ServiceContext) registerConsumers() {
	// 订单超时
	if err := s.MQClient.Subscribe(s.ctx, "order.timeout", s.handleOrderTimeout); err != nil {
		logx.Errorf("order.timeout subscribe failed,error:%v", err.Error())
	}

	// 订单创建（异步下单）
	if err := s.MQClient.Subscribe(s.ctx, constants.ORDER_CREATE_ROUTING_KEY, s.orderCreateHandler); err != nil {
		logx.Errorf("order.create subscribe failed,error:%v", err.Error())
	}
}

// 订单超时处理器
func (s *ServiceContext) handleOrderTimeout(msg []byte) error {
	// 创建一个超时处理上下文
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	var data order2.OrderTimeoutMessage

	if err := json.Unmarshal(msg, &data); err != nil {
		logx.WithContext(ctx).Errorf("解析超时消息失败: %v, raw=%s", err, string(msg))
		return err
	}

	// 查询订单
	order, err := s.OrderInfoModel.FindOneByOrderSn(ctx, data.OrderSn)
	if err != nil {
		logx.WithContext(ctx).Errorf("查询订单失败, order_sn=%s, err=%v", data.OrderSn, err)
		return err
	}

	// 检查订单状态
	if order.Status != constants.ORDER_STATUS_PENDING_PAY {
		logx.WithContext(ctx).Infof("订单状态已变更, 忽略超时处理, order_sn=%s, status=%d", data.OrderSn, order.Status)
		return nil // 直接成功，不重试
	}

	// 1. 先解锁库存（RPC调用放在事务外）
	unlockItems := make([]*productinternal.SkuStockItem, 0)
	for _, item := range data.Items {
		unlockItems = append(unlockItems, &productinternal.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Quantity,
		})
	}
	_, err = s.ProductRpc.UnlockStock(ctx, &productinternal.UpdateStockReq{Items: unlockItems, OrderSn: data.OrderSn})
	if err != nil {
		logx.WithContext(ctx).Errorf("解锁库存失败, order_sn=%s, err=%v", data.OrderSn, err)
		return err
	}

	// 2. 再更新订单状态（本地事务，仅操作数据库）
	err = s.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		order.Status = constants.ORDER_STATUS_CANCELED
		order.CancelTime = sql.NullTime{Time: time.Now(), Valid: true}
		order.CancelReason = data.CancelReason
		order.CancelReasonType = int64(data.CancelReasonType)
		err := s.OrderInfoModel.UpdateTx(ctx, session, order)
		if err != nil {
			return err
		}

		// TODO: 释放优惠券

		return nil
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("取消订单失败, order_sn=%s, err=%v", data.OrderSn, err)
		return err
	}

	logx.WithContext(ctx).Infof("订单超时取消成功, order_sn=%s", data.OrderSn)
	return nil
}
