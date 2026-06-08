package orderinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type NotifyOrderTimeoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotifyOrderTimeoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyOrderTimeoutLogic {
	return &NotifyOrderTimeoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *NotifyOrderTimeoutLogic) NotifyOrderTimeout(in *order.OrderTimeoutNotifyRequest) (*order.Empty, error) {
	l.Logger.Infof("收到订单超时通知,order_sn=%s", in.OrderSn)
	return &order.Empty{}, nil
}
