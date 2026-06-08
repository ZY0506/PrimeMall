package orderinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/constants"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *NotifyRefundSuccessLogic) NotifyRefundSuccess(in *order.RefundSuccessNotifyRequest) (*order.Empty, error) {
	if in.OrderSn == "" {
		l.Logger.Errorf("退款成功通知参数不完整: order_sn=%s", in.OrderSn)
		return &order.Empty{}, nil
	}

	l.Logger.Infof("收到退款成功通知,order_sn=%s,refund_sn=%s", in.OrderSn, in.RefundSn)

	// 查询订单
	orderInfo, err := l.svcCtx.OrderInfoModel.FindOneByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		l.Logger.Errorf("查询订单失败, order_sn=%s, err=%v", in.OrderSn, err)
		return &order.Empty{}, nil
	}

	// 如果订单处于售后中状态，恢复为已完成状态（避免订单永久滞留"售后中"）
	if orderInfo.Status == constants.ORDER_STATUS_AFTER_SALE {
		orderInfo.Status = constants.ORDER_STATUS_COMPLETED
		err = l.svcCtx.OrderInfoModel.Update(l.ctx, orderInfo)
		if err != nil {
			l.Logger.Errorf("恢复订单状态失败, order_sn=%s, err=%v", in.OrderSn, err)
			return &order.Empty{}, nil
		}
		l.Logger.Infof("售后完成，订单状态已恢复: order_sn=%s, status=COMPLETED", in.OrderSn)
	}

	return &order.Empty{}, nil
}
