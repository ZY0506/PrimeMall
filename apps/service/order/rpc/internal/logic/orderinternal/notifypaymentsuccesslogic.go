package orderinternallogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productinternal"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type NotifyPaymentSuccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotifyPaymentSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyPaymentSuccessLogic {
	return &NotifyPaymentSuccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// NotifyPaymentSuccess 处理支付成功通知
// 核心职责：更新订单状态为已支付 + 扣减已锁定库存（幂等安全）
func (l *NotifyPaymentSuccessLogic) NotifyPaymentSuccess(in *order.PaymentSuccessNotifyRequest) (*order.Empty, error) {
	// 1. 参数校验
	if in.OrderSn == "" || in.PaymentSn == "" {
		l.Logger.Errorf("支付通知参数不完整, order_sn=%s, payment_sn=%s", in.OrderSn, in.PaymentSn)
		return nil, errorx.NewBizError(response.ErrCodeInvalidParam, "参数不完整")
	}

	// 2. 查询订单
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		l.Logger.Errorf("查询订单失败, order_sn=%s, err=%v", in.OrderSn, err)
		return nil, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单不存在")
	}

	// 3. 幂等处理：如果已支付则直接成功
	if orderInfo.Status == constants.ORDER_STATUS_PAID {
		l.Logger.Infof("订单已支付, 幂等跳过, order_sn=%s", in.OrderSn)
		return &order.Empty{}, nil
	}

	// 4. 状态校验：仅待支付订单可确认支付
	if orderInfo.Status != constants.ORDER_STATUS_PENDING_PAY {
		l.Logger.Errorf("订单状态不允许支付, order_sn=%s, status=%d", in.OrderSn, orderInfo.Status)
		return nil, errorx.NewBizError(response.ErrCodeOrderStatusInvalid, "订单状态不允许支付")
	}

	// 5. 查询订单商品列表（用于扣减库存）
	items, err := l.svcCtx.OrderItemModel.FindListByOrderId(l.ctx, orderInfo.Id)
	if err != nil {
		l.Logger.Errorf("查询订单商品失败, order_sn=%s, err=%v", in.OrderSn, err)
		return nil, err
	}

	// 6. 组装扣减库存参数
	deductItems := make([]*productinternal.SkuStockItem, 0, len(items))
	for _, item := range items {
		deductItems = append(deductItems, &productinternal.SkuStockItem{
			SkuId:    item.SkuId,
			Quantity: item.Count,
		})
	}

	// 7. 先扣减库存（RPC调用放在事务外）
	_, err = l.svcCtx.ProductRpc.DeductStock(l.ctx, &productinternal.UpdateStockReq{
		Items:   deductItems,
		OrderSn: in.OrderSn,
	})
	if err != nil {
		l.Logger.Errorf("扣减库存失败, order_sn=%s, err=%v", in.OrderSn, err)
		return nil, err
	}

	// 8. 再更新订单状态（本地事务，仅操作数据库）
	var payType int64 = 1 // 默认微信支付（实际按回调参数设置）
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		orderInfo.Status = constants.ORDER_STATUS_PAID
		orderInfo.PayTime = sql.NullTime{Time: time.Now(), Valid: true}
		orderInfo.PayType = payType
		if err := l.svcCtx.OrderInfoModel.UpdateTx(ctx, session, orderInfo); err != nil {
			l.Logger.Errorf("更新订单支付状态失败, order_sn=%s, err=%v", in.OrderSn, err)
			return err
		}
		return nil
	})
	if err != nil {
		l.Logger.Errorf("支付成功处理失败, order_sn=%s, err=%v", in.OrderSn, err)
		return nil, err
	}

	l.Logger.Infof("支付成功处理完成, order_sn=%s, payment_sn=%s", in.OrderSn, in.PaymentSn)
	return &order.Empty{}, nil
}
