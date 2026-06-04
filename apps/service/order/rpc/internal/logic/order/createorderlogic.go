package orderlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateOrder 创建订单（高并发最终版：锁库存移出事务 + 订单项快照）
func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderRequest) (resp *order.CreateOrderResponse, err error) {
	// 1. 获取用户ID
	var userId uint64
	userId, err = ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败，error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeUnauthorized, "登录失效")
	}

	// 2. 幂等性校验
	var ok bool
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.ORDER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等性校验失败: %v", err)
		return nil, err
	}
	if !ok {
		var existOrder *model.OrderInfo
		existOrder, err = l.svcCtx.OrderInfoModel.FindOneByUserIdIdempotencyKey(l.ctx, userId, in.IdempotencyKey)
		if err != nil {
			if !errors.Is(err, model.ErrNotFound) {
				l.Logger.Errorf("查询订单失败: %v", err)
				return nil, err
			}
			// Lua脚本原子操作：删除旧键并重新设置，防止并发竞态
			_, err = l.svcCtx.Client.Eval(l.ctx, constants.LuaResetIdempotencyKey, []string{idempotencyKey}, "1", int64(constants.IDEMPOTENCY_EXIRE/time.Millisecond)).Result()
			if err != nil {
				return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复提交")
			}
		} else {
			return &order.CreateOrderResponse{
				OrderSn:   existOrder.OrderSn,
				PayAmount: existOrder.PayAmount,
			}, nil
		}
	}
	defer func() {
		if err != nil {
			_ = l.svcCtx.Client.Del(l.ctx, idempotencyKey).Err()
		}
	}()

	// 3. 解析缓存（快照数据，无任何查询）
	cacheKey := constants.SETTLEMENT_TOKEN_KEY + in.SettlementToken
	var cacheJson string
	cacheJson, err = l.svcCtx.Client.Get(l.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var settlementData SettlementData
	if err = json.Unmarshal([]byte(cacheJson), &settlementData); err != nil {
		l.Logger.Errorf("解析缓存失败: %v", err)
		return nil, err
	}

	// 权限+金额校验
	if settlementData.UserId != userId {
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "无权限")
	}

	// 校验地址和优惠卷信息
	if settlementData.AddressSnapshot == "" {
		addr, err := l.svcCtx.UserRpc.GetAddressById(l.ctx, &user.GetAddressReq{
			AddressId: in.AddressId,
			UserId:    userId,
		})
		if err != nil {
			l.Logger.Errorf("查询地址失败，error=%v", err)
			return nil, err
		}
		addrSnapshot := &order.AddressSnapshot{
			ReceiverName:  addr.ReceiverName,
			ReceiverPhone: addr.ReceiverPhone,
			Detail: &order.AddressDetail{
				Province:      addr.Address.Province,
				City:          addr.Address.City,
				District:      addr.Address.District,
				DetailAddress: addr.Address.Detail,
				PostalCode:    addr.Address.PostalCode,
			},
		}
		// 序列化地址快照
		addrSnapshotStr, err := json.Marshal(addrSnapshot)
		if err != nil {
			l.Logger.Errorf("序列化地址快照失败，error=%v", err)
			return nil, err
		}
		settlementData.AddressSnapshot = string(addrSnapshotStr)
	}
	if settlementData.CouponId != 0 {
		// TODO: 校验优惠卷信息是否正确、计算优惠卷扣减金额
		settlementData.CouponId = in.CouponId
	}

	// 4. 生成订单ID + 订单号
	orderId, _ := l.svcCtx.IDGenerator.NextID()
	orderSn, _ := l.svcCtx.IDGenerator.GenWithPrefix(constants.PREFIX_ORDER_SN)
	now := time.Now()

	// 5. 构建库存参数 + 数据库订单项
	length := len(settlementData.ItemSnapshots)
	lockStockItems := make([]*product.SkuStockItem, 0, length)
	orderItems := make([]*model.OrderItem, 0, length)
	unlockStockItems := make([]*order.OrderItemSimple, 0, length)
	for _, s := range settlementData.ItemSnapshots {
		// 库存锁定参数
		lockStockItems = append(lockStockItems, &product.SkuStockItem{
			SkuId:    s.SkuId,
			Quantity: s.Count,
		})
		// 超时取消订单参数（解锁库存）
		unlockStockItems = append(unlockStockItems, &order.OrderItemSimple{
			SkuId:    s.SkuId,
			Quantity: s.Count,
		})
		// 订单项ID
		itemId, _ := l.svcCtx.IDGenerator.NextID()
		orderItems = append(orderItems, &model.OrderItem{
			Id:          itemId,
			OrderId:     orderId,
			OrderSn:     orderSn,
			SkuId:       s.SkuId,
			SpuId:       s.SpuId,
			SpuName:     s.SpuName,
			SkuName:     s.SkuName,
			SkuPic:      s.SkuPic,
			Price:       s.Price,
			Count:       s.Count,
			TotalAmount: s.TotalAmount,
			IsSeckill:   0,
			CreatedAt:   now,
		})
	}

	// 6. 构建订单主表
	orderInfo := &model.OrderInfo{
		Id:             orderId,
		OrderSn:        orderSn,
		UserId:         userId,
		OrderType:      constants.ORDER_TYPE_NORMAL,
		Status:         constants.ORDER_STATUS_PENDING_PAY,
		PayType:        int64(in.PayType),
		PayAmount:      settlementData.PayAmount,
		TotalAmount:    settlementData.TotalAmount,
		FreightAmount:  settlementData.FreightAmount,
		CouponId:       settlementData.CouponId,
		CouponDiscount: settlementData.CouponDiscount,
		Remark:         in.Remark,
		AddressSnap:    settlementData.AddressSnapshot,
		IdempotencyKey: in.IdempotencyKey,
		CreatedAt:      now,
		ExpireTime:     sql.NullTime{Time: now.Add(constants.ORDER_EXPIRE_TIME), Valid: true},
		UpdatedAt:      now,
	}

	// ===================== 锁库存  =====================
	lockRet, err := l.svcCtx.ProductRpc.LockStock(l.ctx, &product.UpdateStockReq{
		Items:   lockStockItems,
		OrderSn: orderSn,
	})
	if err != nil || !lockRet.Success {
		var msg string
		if err != nil {
			msg = "库存锁定失败"
		} else {
			var msgs []string
			for _, v := range lockRet.Results {
				if !v.Success {
					msgs = append(msgs, fmt.Sprintf("SKU:%d %s", v.SkuId, v.Message))
				}
			}
			msg = strings.Join(msgs, "；")
		}
		l.Logger.Errorf("锁库存失败: %s,error=%v", msg, err)
		return nil, errorx.NewBizError(response.ErrCodeOrderFailed, msg)
	}

	// TODO:锁优惠卷
	// 检查优惠卷状态
	// 锁优惠卷、记录优惠卷使用和订单信息
	// 支付成功再扣减优惠卷

	// 7. 本地事务：仅做数据库插入
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 插入订单
		_, err = l.svcCtx.OrderInfoModel.InsertTx(l.ctx, session, orderInfo)
		if err != nil {
			return err
		}
		// 插入订单项
		for _, item := range orderItems {
			_, err = l.svcCtx.OrderItemModel.InsertTx(l.ctx, session, item)
			if err != nil {
				return err
			}
		}
		return nil
	})
	// 事务失败：解锁库存
	if err != nil {
		_, _ = l.svcCtx.ProductRpc.UnlockStock(l.ctx, &product.UpdateStockReq{
			Items:   lockStockItems,
			OrderSn: orderSn,
		})
		l.Logger.Errorf("事务失败: %v", err)
		return nil, err
	}

	// 删除缓存
	_ = l.svcCtx.Client.Del(l.ctx, cacheKey).Err()

	// 8. 后置处理
	if len(in.CartSkuIds) > 0 {
		_ = l.svcCtx.CartModel.BatchDelete(l.ctx, userId, in.CartSkuIds)
	}
	_ = l.svcCtx.Client.Expire(l.ctx, idempotencyKey, 24*time.Hour).Err()
	l.Logger.Info("创建订单成功")

	// 发送消息到延迟队列，处理订单超时；
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.Logger.Errorf("发送超时消息协程panic: %v", r)
			}
		}()
		msg, _ := json.Marshal(&order.OrderTimeoutMessage{
			OrderSn:          orderSn,
			CancelReason:     "超时取消",
			CancelReasonType: constants.CANCEL_REASON_TYPE_TIMEOUT,
			Items:            unlockStockItems,
			CouponId:         in.CouponId,
		})
		err = l.svcCtx.MQClient.Publish(l.ctx, "", constants.ORDER_TIMEOUT_ROUTING_KEY, msg)
		if err != nil {
			l.Logger.Errorf("发送消息失败: %v", err)
			return
		}
	}()

	return &order.CreateOrderResponse{
		OrderSn:   orderSn,
		PayAmount: orderInfo.PayAmount,
	}, nil
}
