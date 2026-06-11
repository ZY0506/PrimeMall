package orderlogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/client/marketing"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelOrder 取消订单
func (l *CancelOrderLogic) CancelOrder(in *order.CancelOrderRequest) (*order.Empty, error) {
	// 查询用户信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	// 查找订单信息
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("订单号：%s 不存在", in.OrderSn)
			return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
		}
		l.Logger.Errorf("获取订单信息失败,error=%v", err)
		return nil, err
	}
	if orderInfo.UserId != userId {
		l.Logger.Errorf("用户ID不匹配,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodePermissionDenied, "用户无权限操作")
	}
	// 校验状态
	if orderInfo.Status != constants.ORDER_STATUS_PENDING_PAY {
		l.Logger.Errorf("订单状态错误,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeOrderCancelFailed, "订单取消失败")
	}

	// 检查支付截止时间是否已过
	if !orderInfo.ExpireTime.Valid || time.Now().After(orderInfo.ExpireTime.Time) {
		l.Logger.Errorf("订单已超时取消，orderSn=%v", orderInfo.OrderSn)
		return nil, errorx.NewBizError(response.ErrCodeOrderExpired, "订单已超时取消")
	}
	// 查询该订单的所有商品项（用于恢复库存）
	orderItems, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, orderInfo.Id)
	if err != nil {
		return nil, err
	}

	// 构建需要解锁的库存列表
	items := make([]*product.SkuStockItem, 0, len(orderItems))
	for _, item := range orderItems {
		items = append(items, &product.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Count,
		})
	}

	// 1. 先解锁库存（RPC调用放在事务外，避免分布式事务问题）
	unlockRet, err := l.svcCtx.ProductRpc.UnlockStock(l.ctx, &product.UpdateStockReq{
		OrderSn: in.OrderSn,
		Items:   items,
	})
	if err != nil {
		logx.Errorf("解锁库存RPC调用失败: %v", err)
		return nil, errorx.NewBizError(response.ErrCodeOrderCancelFailed, "取消失败，请重试")
	}
	if !unlockRet.Success {
		var msgs []string
		for _, v := range unlockRet.Results {
			if !v.Success {
				msgs = append(msgs, fmt.Sprintf("SKU:%d %s", v.SkuId, v.Message))
			}
		}
		return nil, errorx.NewBizError(response.ErrCodeOrderCancelFailed, strings.Join(msgs, ";"))
	}

	// 1.1 释放优惠券（如果使用了优惠券）
	if orderInfo.CouponId > 0 {
		ctxWithUid, ctxErr := ctxdata.PutUserIdToCtx(l.ctx, userId)
		if ctxErr == nil {
			l.unlockCouponWithRetry(ctxWithUid, orderInfo.CouponId, orderInfo.OrderSn)
		}
	}

	// 2. 开启数据库事务（仅更新订单状态等本地资源）
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新订单状态为"已取消"，并记录取消时间
		orderInfo.Status = constants.ORDER_STATUS_CANCELED
		orderInfo.CancelTime = sql.NullTime{Time: time.Now(), Valid: true}
		orderInfo.CancelReasonType = int64(in.ReasonType)
		orderInfo.CancelReason = in.Reason
		err = l.svcCtx.OrderInfoModel.UpdateTx(l.ctx, session, orderInfo)
		if err != nil {
			l.Logger.Errorf("更新订单状态失败,error=%v", err)
			return err
		}
		return nil
	})

	if err != nil {
		l.Logger.Errorf("（事务）取消订单失败,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeOrderCancelFailed, "取消失败，请重试")
	}

	l.Logger.Info("取消订单成功")
	return &order.Empty{}, nil
}

// unlockCouponWithRetry 解锁优惠券，带重试机制（尽力而为）
func (l *CancelOrderLogic) unlockCouponWithRetry(ctx context.Context, userCouponId uint64, orderSn string) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(100*(1<<attempt)) * time.Millisecond)
		}
		uCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		resp, err := l.svcCtx.MarketingRpc.UnlockCoupon(uCtx, &marketing.UnlockCouponReq{
			UserCouponId: userCouponId,
			OrderSn:      orderSn,
		})
		cancel()
		if err == nil && resp != nil && resp.Success {
			l.Logger.Infof("解锁优惠券成功, user_coupon_id=%d", userCouponId)
			return
		}
		lastErr = err
		if resp != nil {
			l.Logger.Errorf("解锁优惠券失败（第%d次）, user_coupon_id=%d, reason=%s",
				attempt+1, userCouponId, resp.ErrorMsg)
		} else {
			l.Logger.Errorf("解锁优惠券RPC失败（第%d次）, user_coupon_id=%d, err=%v",
				attempt+1, userCouponId, err)
		}
	}
	if lastErr != nil {
		l.Logger.Errorf("解锁优惠券最终失败（需人工处理）, user_coupon_id=%d, order_sn=%s",
			userCouponId, orderSn)
	}
}
