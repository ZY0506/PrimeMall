package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/client/marketing"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
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

// orderCreateHandler 处理订单创建消息：锁库存 → 使用优惠券 → 落库 → 后置处理
func (sc *ServiceContext) orderCreateHandler(msg []byte) error {
	ctx, cancel := context.WithTimeout(sc.Ctx, 10*time.Second)
	defer cancel()
	var createMsg OrderCreateMessage
	if err := json.Unmarshal(msg, &createMsg); err != nil {
		logx.WithContext(ctx).Errorf("解析订单创建消息失败: %v, raw=%s", err, string(msg))
		return err
	}

	logx.WithContext(ctx).Infof("开始异步创建订单，order_sn=%s", createMsg.OrderSn)

	// ===================== 1. 幂等校验 =====================
	// 1a. 按 OrderSn 检查（同一次请求重试）
	exists, err := sc.OrderInfoModel.FindOneByOrderSn(ctx, createMsg.OrderSn)
	if err == nil && exists != nil {
		logx.WithContext(ctx).Infof("订单已存在（幂等），order_sn=%s", createMsg.OrderSn)
		return nil
	}
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return err
	}

	// 1b. 按 (user_id, idempotency_key) 检查（相同幂等键的旧订单已存在）
	exists, err = sc.OrderInfoModel.FindOneByUserIdIdempotencyKey(ctx, createMsg.UserId, createMsg.IdempotencyKey)
	if err == nil && exists != nil {
		logx.WithContext(ctx).Infof("订单已存在（幂等，user_id+idempotency_key），order_sn=%s", createMsg.OrderSn)
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

	// ===================== 5. 使用优惠券（下单前扣减，保证一致性）=====================
	usedCoupon := false
	if createMsg.CouponId > 0 {
		useCouponCtx := ctx
		if ckCtx, ckErr := ctxdata.PutUserIdToCtx(ctx, createMsg.UserId); ckErr == nil {
			useCouponCtx = ckCtx
		} else {
			logx.WithContext(ctx).Errorf("注入用户ID到优惠券上下文失败, order_sn=%s, err=%v", createMsg.OrderSn, ckErr)
		}
		useRet, useErr := sc.useCouponWithRetry(useCouponCtx, &createMsg)
		if useErr != nil || useRet == nil || !useRet.Success {
			errMsg := "优惠券使用失败"
			if useRet != nil {
				errMsg = useRet.ErrorMsg
			}
			logx.WithContext(ctx).Errorf("使用优惠券失败，order_sn=%s, reason=%s", createMsg.OrderSn, errMsg)
			// 补偿：解锁库存
			_, _ = sc.ProductRpc.UnlockStock(ctx, &product.UpdateStockReq{
				Items:   lockStockItems,
				OrderSn: createMsg.OrderSn,
			})
			return sc.handleFailOrder(ctx, orderInfo, now, "优惠券使用失败，订单自动取消", constants.CANCEL_REASON_TYPE_OTHER)
		}
		usedCoupon = true
		// 用 UseCoupon 返回的实际折扣金额覆盖 payAmount
		if useRet.DiscountAmount > 0 {
			actualPay := createMsg.TotalAmount + createMsg.FreightAmount - useRet.DiscountAmount
			if actualPay < 0 {
				actualPay = 0
			}
			orderInfo.PayAmount = actualPay
			orderInfo.CouponDiscount = useRet.DiscountAmount
		}
	}

	// ===================== 6. 本地事务：落库 =====================
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
		// 事务失败：解锁库存 + 解锁优惠券（补偿）
		_, _ = sc.ProductRpc.UnlockStock(ctx, &product.UpdateStockReq{
			Items:   lockStockItems,
			OrderSn: createMsg.OrderSn,
		})
		if usedCoupon {
			sc.unlockCouponBestEffort(ctx, &createMsg)
		}
		logx.WithContext(ctx).Errorf("异步订单事务失败，order_sn=%s, err=%v", createMsg.OrderSn, err)
		return err
	}

	// ===================== 7. 后置处理 =====================
	// 自动清除已购买商品的购物车记录
	cartSkuIds := createMsg.CartSkuIds
	if len(cartSkuIds) == 0 {
		// 未传入 CartSkuIds 时，从订单商品 SKU 列表自动推导
		for _, item := range createMsg.ItemSnapshots {
			cartSkuIds = append(cartSkuIds, item.SkuId)
		}
	}
	if len(cartSkuIds) > 0 {
		_ = sc.CartModel.BatchDelete(ctx, createMsg.UserId, cartSkuIds)
	}

	// 发送延迟队列 → 超时取消
	timeoutMsg, _ := json.Marshal(&order.OrderTimeoutMessage{
		OrderSn:          createMsg.OrderSn,
		CancelReason:     "超时取消",
		CancelReasonType: constants.CANCEL_REASON_TYPE_TIMEOUT,
		Items:            unlockStockItems,
		CouponId:         createMsg.CouponId,
	})
	_ = sc.MQClient.Publish(ctx, constants.ORDER_DELAY_QUEUE+"_exchange", constants.ORDER_DELAY_QUEUE, timeoutMsg)

	logx.WithContext(ctx).Infof("异步订单创建成功，order_sn=%s", createMsg.OrderSn)
	return nil
}

// useCouponWithRetry 调用营销服务使用优惠券，带重试机制
// 仅对网络等瞬时错误重试，业务错误（已使用、已过期等）直接返回
func (sc *ServiceContext) useCouponWithRetry(ctx context.Context, createMsg *OrderCreateMessage) (*marketing.UseCouponResp, error) {
	var lastErr error
	maxAttempts := 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// 指数退避：200ms, 400ms
			backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		useCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		resp, err := sc.MarketingRpc.UseCoupon(useCtx, &marketing.UseCouponReq{
			UserCouponId: createMsg.CouponId,
			OrderSn:      createMsg.OrderSn,
			OrderAmount:  createMsg.TotalAmount,
		})
		cancel()

		if err != nil {
			// 网络/超时等瞬时错误，可重试
			lastErr = err
			logx.WithContext(ctx).Errorf("使用优惠券RPC失败（第%d次），order_sn=%s, err=%v",
				attempt+1, createMsg.OrderSn, err)
			continue
		}

		if !resp.Success {
			// 业务错误（已使用、已过期等），不可重试
			return resp, nil
		}

		// 成功
		return resp, nil
	}

	return nil, fmt.Errorf("使用优惠券重试%d次后仍然失败: %v", maxAttempts, lastErr)
}

// unlockCouponBestEffort 尽力解锁优惠券（补偿操作），不返回错误避免打断主流程
func (sc *ServiceContext) unlockCouponBestEffort(ctx context.Context, createMsg *OrderCreateMessage) {
	if createMsg.CouponId == 0 {
		return
	}
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
		}
		uCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		resp, err := sc.MarketingRpc.UnlockCoupon(uCtx, &marketing.UnlockCouponReq{
			UserCouponId: createMsg.CouponId,
			OrderSn:      createMsg.OrderSn,
		})
		cancel()
		if err == nil && resp != nil && resp.Success {
			return
		}
		logx.WithContext(ctx).Errorf("解锁优惠券失败（第%d次），order_sn=%s, err=%v",
			attempt+1, createMsg.OrderSn, err)
	}
	logx.WithContext(ctx).Errorf("解锁优惠券最终失败，需人工处理，order_sn=%s, coupon_id=%d",
		createMsg.OrderSn, createMsg.CouponId)
}

// handleStockFailOrder 库存不足：记录订单为取消状态
func (sc *ServiceContext) handleStockFailOrder(ctx context.Context, orderInfo *model.OrderInfo, now time.Time) error {
	return sc.handleFailOrder(ctx, orderInfo, now, "库存不足，订单自动取消", constants.CANCEL_REASON_TYPE_INSUFFICIENT_STOCK)
}

// handleFailOrder 记录订单为取消状态
func (sc *ServiceContext) handleFailOrder(ctx context.Context, orderInfo *model.OrderInfo, now time.Time, reason string, reasonType int64) error {
	orderInfo.Status = constants.ORDER_STATUS_CANCELED
	orderInfo.CancelTime = sql.NullTime{Time: now, Valid: true}
	orderInfo.CancelReason = reason
	orderInfo.CancelReasonType = reasonType

	return sc.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		_, err := sc.OrderInfoModel.InsertTx(ctx, session, orderInfo)
		return err
	})
}
