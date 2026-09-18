package orderinternallogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productinternal"
	"github.com/ZY0506/PrimeMall/common/constants"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type NotifyRefundSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotifyRefundSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyRefundSuccessLogic {
	return &NotifyRefundSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// NotifyRefundSuccess 处理退款成功通知。
// 核心职责：认领退款状态 → 按售后类型回补库存 → 恢复订单状态。
//
// 幂等由两把互不依赖的锁保证：
//   - after_sale.refund_status 的条件原子更新（CAS）保证「认领恰好一次」；
//   - stock_log 以退款单号为键、且 RollbackStock 开启了幂等校验，保证「回补恰好一次」。
//
// 因此 CAS 命中 0 行不能提前返回：那可能表示上一轮通知已认领、但回补之前进程挂了。
func (l *NotifyRefundSuccessLogic) NotifyRefundSuccess(in *order.RefundSuccessNotifyRequest) (*order.Empty, error) {
	if in.AfterSaleId == 0 || in.OrderSn == "" || in.RefundSn == "" {
		l.Logger.Errorf("退款成功通知参数不完整: order_sn=%s, refund_sn=%s, after_sale_id=%d",
			in.OrderSn, in.RefundSn, in.AfterSaleId)
		return &order.Empty{}, nil
	}

	l.Logger.Infof("收到退款成功通知, order_sn=%s, refund_sn=%s, after_sale_id=%d",
		in.OrderSn, in.RefundSn, in.AfterSaleId)

	// 1. 载入售后单，并校验它与通知里的订单号一致（防止串单）
	afterSale, err := l.svcCtx.AfterSaleModel.FindOne(l.ctx, in.AfterSaleId)
	if err != nil {
		l.Logger.Errorf("查询售后单失败, after_sale_id=%d, err=%v", in.AfterSaleId, err)
		return &order.Empty{}, nil
	}
	if afterSale.OrderSn != in.OrderSn {
		l.Logger.Errorf("售后单与订单号不匹配, after_sale_id=%d, 售后单order_sn=%s, 通知order_sn=%s",
			in.AfterSaleId, afterSale.OrderSn, in.OrderSn)
		return &order.Empty{}, nil
	}

	// 2. CAS 认领退款状态：退款中 → 退款成功
	claimed, err := l.svcCtx.AfterSaleModel.CasRefundStatus(l.ctx,
		afterSale.Id, constants.REFUND_STATUS_REFUNDING, constants.REFUND_STATUS_REFUNDED)
	if err != nil {
		l.Logger.Errorf("更新售后单退款状态失败, after_sale_id=%d, err=%v", afterSale.Id, err)
		return nil, err
	}
	if claimed == 0 {
		l.Logger.Infof("退款状态未被本次认领, 继续幂等回补, after_sale_id=%d, refund_status=%d",
			afterSale.Id, afterSale.RefundStatus)
	}

	// 3. 回补库存
	if err := l.restockForReturnRefund(afterSale.Type, afterSale.AfterSaleSn, afterSale.OrderSn, in.RefundSn); err != nil {
		return nil, err
	}

	// 4. 恢复订单状态，避免订单永久滞留「售后中」
	if err := l.restoreOrderStatus(in.OrderSn); err != nil {
		return nil, err
	}

	return &order.Empty{}, nil
}

// restockForReturnRefund 按售后类型决定是否回补库存。
//
// 只有「退货退款」（type=2）才回补：「仅退款」（type=1）的货物仍在买家手上，
// 回补等于凭空造出库存，是与本次要修的缺陷同一类的超卖来源。
func (l *NotifyRefundSuccessLogic) restockForReturnRefund(afterSaleType int64, afterSaleSn, orderSn, refundSn string) error {
	if afterSaleType != constants.AFTER_SALE_TYPE_RETURN_REFUND {
		l.Logger.Infof("售后类型为仅退款, 不回补库存, after_sale_sn=%s, type=%d", afterSaleSn, afterSaleType)
		return nil
	}

	item, err := l.svcCtx.AfterSaleItemModel.FindOneByAfterSaleSn(l.ctx, afterSaleSn)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			// 退货退款单必然有商品快照，缺了属于数据问题，重试也不会自愈
			l.Logger.Errorf("退货退款售后单缺少商品快照, 无法回补库存, after_sale_sn=%s", afterSaleSn)
			return nil
		}
		l.Logger.Errorf("查询售后商品快照失败, after_sale_sn=%s, err=%v", afterSaleSn, err)
		return err
	}
	if item.Quantity <= 0 {
		l.Logger.Errorf("售后商品数量非法, 跳过回补, after_sale_sn=%s, quantity=%d", afterSaleSn, item.Quantity)
		return nil
	}

	// 幂等键用退款单号而非原订单号：stock_log 的唯一键是 (order_sn, sku_id, change_type)，
	// 而一个订单可以发生多次部分退款，用订单号做键会让第二次退款被幂等判据静默跳过。
	_, err = l.svcCtx.ProductRpc.RollbackStock(l.ctx, &productinternal.UpdateStockReq{
		Items: []*productinternal.SkuStockItem{
			{SkuId: item.SkuId, Quantity: item.Quantity},
		},
		OrderSn: refundSn,
	})
	if err != nil {
		l.Logger.Errorf("回补库存失败, order_sn=%s, refund_sn=%s, sku_id=%d, quantity=%d, err=%v",
			orderSn, refundSn, item.SkuId, item.Quantity, err)
		return err
	}

	l.Logger.Infof("回补库存成功, order_sn=%s, refund_sn=%s, sku_id=%d, quantity=%d",
		orderSn, refundSn, item.SkuId, item.Quantity)
	return nil
}

// restoreOrderStatus 把订单从「售后中」恢复到「已完成」。
func (l *NotifyRefundSuccessLogic) restoreOrderStatus(orderSn string) error {
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, orderSn)
	if err != nil {
		l.Logger.Errorf("查询订单失败, order_sn=%s, err=%v", orderSn, err)
		return nil
	}
	if orderInfo.Status != constants.ORDER_STATUS_AFTER_SALE {
		return nil
	}

	orderInfo.Status = constants.ORDER_STATUS_COMPLETED
	if err := l.svcCtx.OrderInfoModel.Update(l.ctx, orderInfo); err != nil {
		l.Logger.Errorf("恢复订单状态失败, order_sn=%s, err=%v", orderSn, err)
		return err
	}
	l.Logger.Infof("售后完成，订单状态已恢复: order_sn=%s, status=COMPLETED", orderSn)
	return nil
}
