package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// OrderCreateMessage 异步订单创建消息（扁平结构，避免循环依赖）
type OrderCreateMessage struct {
	OrderId         uint64   `json:"order_id"`
	OrderSn         string   `json:"order_sn"`
	UserId          uint64   `json:"user_id"`
	SettlementToken string   `json:"settlement_token"`
	PayType         int64    `json:"pay_type"`
	Remark          string   `json:"remark"`
	IdempotencyKey  string   `json:"idempotency_key"`
	CartSkuIds      []uint64 `json:"cart_sku_ids"`
	CouponId        uint64   `json:"coupon_id"`
	// SettlementData 扁平化字段（原 SettlementData 的所有字段）
	AddressSnapshot string             `json:"address_snapshot"`
	ItemSnapshots   []OrderItemMessage `json:"item_snapshots"`
	TotalAmount     int64              `json:"total_amount"`
	FreightAmount   int64              `json:"freight_amount"`
	CouponDiscount  int64              `json:"coupon_discount"`
	PayAmount       int64              `json:"pay_amount"`
}

// OrderItemMessage 订单项快照（用于消息传递）
type OrderItemMessage struct {
	SkuId       uint64 `json:"sku_id"`
	SpuId       uint64 `json:"spu_id"`
	SpuName     string `json:"spu_name"`
	SkuName     string `json:"sku_name"`
	SkuPic      string `json:"sku_pic"`
	Price       int64  `json:"price"`
	Count       int64  `json:"count"`
	TotalAmount int64  `json:"total_amount"`
}

// orderCreateHandler 处理订单创建消息：锁库存 + 落库 + 发送超时队列
func (sc *ServiceContext) orderCreateHandler(msg []byte) error {
	ctx, cancel := context.WithTimeout(sc.ctx, 10*time.Second)
	defer cancel()
	var createMsg OrderCreateMessage
	if err := json.Unmarshal(msg, &createMsg); err != nil {
		logx.WithContext(ctx).Errorf("解析订单创建消息失败: %v, raw=%s", err, string(msg))
		return err
	}

	logx.WithContext(ctx).Infof("开始异步创建订单，order_sn=%s", createMsg.OrderSn)

	// ===================== 1. 幂等校验 =====================
	exists, err := sc.OrderInfoModel.FindOneByOrderSn(ctx, createMsg.OrderSn)
	if err == nil && exists != nil {
		logx.WithContext(ctx).Infof("订单已存在（幂等），order_sn=%s", createMsg.OrderSn)
		return nil
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}

	// ===================== 2. 构建库存参数 + 订单项 =====================
	length := len(createMsg.ItemSnapshots)
	lockStockItems := make([]*product.SkuStockItem, 0, length)
	orderItems := make([]*model.OrderItem, 0, length)
	unlockStockItems := make([]*order.OrderItemSimple, 0, length)
	now := time.Now()

	for _, s := range createMsg.ItemSnapshots {
		lockStockItems = append(lockStockItems, &product.SkuStockItem{
			SkuId:    s.SkuId,
			Quantity: s.Count,
		})
		unlockStockItems = append(unlockStockItems, &order.OrderItemSimple{
			SkuId:    s.SkuId,
			Quantity: s.Count,
		})
		itemId, _ := sc.IDGenerator.NextID()
		orderItems = append(orderItems, &model.OrderItem{
			Id:          itemId,
			OrderId:     createMsg.OrderId,
			OrderSn:     createMsg.OrderSn,
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

	// ===================== 3. 构建订单主表 =====================
	orderInfo := &model.OrderInfo{
		Id:             createMsg.OrderId,
		OrderSn:        createMsg.OrderSn,
		UserId:         createMsg.UserId,
		OrderType:      constants.ORDER_TYPE_NORMAL,
		Status:         constants.ORDER_STATUS_PENDING_PAY,
		PayType:        createMsg.PayType,
		PayAmount:      createMsg.PayAmount,
		TotalAmount:    createMsg.TotalAmount,
		FreightAmount:  createMsg.FreightAmount,
		CouponId:       createMsg.CouponId,
		CouponDiscount: createMsg.CouponDiscount,
		Remark:         createMsg.Remark,
		AddressSnap:    createMsg.AddressSnapshot,
		IdempotencyKey: createMsg.IdempotencyKey,
		CreatedAt:      now,
		ExpireTime:     sql.NullTime{Time: now.Add(constants.ORDER_EXPIRE_TIME), Valid: true},
		UpdatedAt:      now,
	}

	// ===================== 4. 锁库存 =====================
	lockRet, err := sc.ProductRpc.LockStock(ctx, &product.UpdateStockReq{
		Items:   lockStockItems,
		OrderSn: createMsg.OrderSn,
	})
	if err != nil || !lockRet.Success {
		logx.WithContext(ctx).Errorf("异步锁库存失败，order_sn=%s, err=%v", createMsg.OrderSn, err)
		return sc.handleStockFailOrder(ctx, orderInfo, now)
	}

	// ===================== 5. 本地事务：落库 =====================
	err = sc.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err := sc.OrderInfoModel.InsertTx(ctx, session, orderInfo)
		if err != nil {
			return err
		}
		for _, item := range orderItems {
			_, err = sc.OrderItemModel.InsertTx(ctx, session, item)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// 事务失败：解锁库存
		_, _ = sc.ProductRpc.UnlockStock(ctx, &product.UpdateStockReq{
			Items:   lockStockItems,
			OrderSn: createMsg.OrderSn,
		})
		logx.WithContext(ctx).Errorf("异步订单事务失败，order_sn=%s, err=%v", createMsg.OrderSn, err)
		return err
	}

	// ===================== 6. 后置处理 =====================
	if len(createMsg.CartSkuIds) > 0 {
		_ = sc.CartModel.BatchDelete(ctx, createMsg.UserId, createMsg.CartSkuIds)
	}

	// 发送延迟队列 → 超时取消
	timeoutMsg, _ := json.Marshal(&order.OrderTimeoutMessage{
		OrderSn:          createMsg.OrderSn,
		CancelReason:     "超时取消",
		CancelReasonType: constants.CANCEL_REASON_TYPE_TIMEOUT,
		Items:            unlockStockItems,
		CouponId:         createMsg.CouponId,
	})
	_ = sc.MQClient.Publish(ctx, "", constants.ORDER_TIMEOUT_ROUTING_KEY, timeoutMsg)

	logx.WithContext(ctx).Infof("异步订单创建成功，order_sn=%s", createMsg.OrderSn)
	return nil
}

// handleStockFailOrder 库存不足：记录订单为取消状态
func (sc *ServiceContext) handleStockFailOrder(ctx context.Context, orderInfo *model.OrderInfo, now time.Time) error {
	orderInfo.Status = constants.ORDER_STATUS_CANCELED
	orderInfo.CancelTime = sql.NullTime{Time: now, Valid: true}
	orderInfo.CancelReason = "库存不足，订单自动取消"
	orderInfo.CancelReasonType = constants.CANCEL_REASON_TYPE_INSUFFICIENT_STOCK

	return sc.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err := sc.OrderInfoModel.InsertTx(ctx, session, orderInfo)
		return err
	})
}
