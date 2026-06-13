package svc

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productinternal"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// isCircuitBreakerError 判断是否是熔断错误
func isCircuitBreakerError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "circuit breaker is open")
}

const (
	closeOrderInterval = 1 * time.Minute // 定时扫描间隔
	closeOrderBatch    = 1000            // 每批处理数量
)

// CloseExpiredOrders 定时扫描并关闭过期未支付订单（兜底方案）
// 当 RabbitMQ 延迟消息丢失时（如系统重启），此任务确保过期订单仍被关闭
func (s *ServiceContext) CloseExpiredOrders(ctx context.Context) {
	logx.Info("定时关单任务启动")

	// 延迟10秒再执行首次扫描，等所有服务注册到 etcd 并就绪
	select {
	case <-ctx.Done():
		logx.Info("定时关单任务退出")
		return
	case <-time.After(10 * time.Second):
	}

	// 启动时立即执行多次兜底扫描（快速清理积压订单）
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			logx.Info("定时关单任务退出")
			return
		default:
		}
		remaining := s.batchCloseExpired(ctx)
		if remaining == 0 {
			break
		}
		logx.Infof("定时关单：首轮批量 %d/10 完成，仍有 %d 个待处理", i+1, remaining)
	}

	ticker := time.NewTicker(closeOrderInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logx.Info("定时关单任务退出")
			return
		case <-ticker.C:
			s.batchCloseExpired(ctx)
		}
	}
}

// batchCloseExpired 批量关闭过期订单，返回剩余待处理数
func (s *ServiceContext) batchCloseExpired(ctx context.Context) int {
	orders, err := s.OrderInfoModel.FindExpiredOrders(ctx, closeOrderBatch)
	if err != nil {
		logx.Errorf("定时关单：查询过期订单失败, error=%v", err)
		return 0
	}
	if len(orders) == 0 {
		return 0
	}

	logx.Infof("定时关单：发现 %d 个过期待支付订单", len(orders))

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return 0
		default:
		}

		if err := s.closeOneExpiredOrder(ctx, order); err != nil {
			logx.Errorf("定时关单：关闭订单失败, order_sn=%s, error=%v", order.OrderSn, err)
			// 如果熔断器已开，停止本轮处理，等下一轮重试
			if isCircuitBreakerError(err) {
				logx.Infof("定时关单：熔断器已开，停止本轮处理")
				return 0
			}
		}
	}
	if len(orders) >= closeOrderBatch {
		return closeOrderBatch
	}
	return 0
}

// AutoCancelExpiredOrder 对外暴露：根据订单号自动取消过期订单
// 供订单详情/列表懒检查调用，返回 true 表示已自动取消
func (s *ServiceContext) AutoCancelExpiredOrder(ctx context.Context, orderSn string) bool {
	order, err := s.OrderInfoModel.FindOneByOrderSn(ctx, orderSn)
	if err != nil {
		return false
	}
	if err := s.closeOneExpiredOrder(ctx, order); err != nil {
		logx.Errorf("懒检查自动关单失败, order_sn=%s, error=%v", orderSn, err)
		return false
	}
	return true
}

// closeOneExpiredOrder 关闭单个过期订单（逻辑与 handleOrderTimeout 一致）
func (s *ServiceContext) closeOneExpiredOrder(ctx context.Context, order *model.OrderInfo) error {
	subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 再次校验订单状态（避免并发修改）
	if order.Status != constants.ORDER_STATUS_PENDING_PAY {
		return nil
	}
	if !order.ExpireTime.Valid || time.Now().Before(order.ExpireTime.Time) {
		return nil
	}

	// 查询订单商品项（用于解锁库存）
	orderItems, err := s.OrderItemModel.FindListByOrderId(subCtx, order.Id)
	if err != nil {
		return err
	}

	// 1. 解锁库存（尽力而为——SKU可能已被删除导致失败，不影响关单）
	unlockItems := make([]*productinternal.SkuStockItem, 0, len(orderItems))
	for _, item := range orderItems {
		unlockItems = append(unlockItems, &productinternal.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Count,
		})
	}
	if _, err := s.ProductRpc.UnlockStock(subCtx, &productinternal.UpdateStockReq{
		Items:   unlockItems,
		OrderSn: order.OrderSn,
	}); err != nil {
		logx.Errorf("定时关单：解锁库存失败（继续关单）, order_sn=%s, error=%v", order.OrderSn, err)
	}

	// 2. 释放优惠券（如果有，尽力而为）
	if order.CouponId > 0 {
		uidCtx, uidErr := ctxdata.PutUserIdToCtx(subCtx, order.UserId)
		if uidErr == nil {
			s.unlockCouponWithRetry(uidCtx, order.CouponId, order.OrderSn)
		}
	}

	// 3. 更新订单状态为已取消
	err = s.DB.TransactCtx(subCtx, func(ctx context.Context, session sqlx.Session) error {
		order.Status = constants.ORDER_STATUS_CANCELED
		order.CancelTime = sql.NullTime{Time: time.Now(), Valid: true}
		order.CancelReasonType = 3 // 超时取消
		order.CancelReason = "订单超时未支付，系统自动取消"
		return s.OrderInfoModel.UpdateTx(ctx, session, order)
	})
	if err != nil {
		logx.Errorf("定时关单：更新订单状态失败, order_sn=%s, error=%v", order.OrderSn, err)
		return err
	}

	logx.Infof("定时关单：订单超时取消成功, order_sn=%s", order.OrderSn)
	return nil
}
