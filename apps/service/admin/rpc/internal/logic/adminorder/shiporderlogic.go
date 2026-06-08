package adminorderlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShipOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ShipOrderLogic) ShipOrder(in *admin.AdminShipOrderReq) (*admin.Empty, error) {
	_, err := l.svcCtx.OrderRpc.ShipOrder(l.ctx, &ordertypes.AdminShipOrderRequest{
		OrderSn:         in.OrderSn,
		DeliveryCompany: in.DeliveryCompany,
		DeliverySn:      in.DeliverySn,
	})
	if err != nil {
		l.Logger.Errorf("订单发货失败,order_sn=%s,error=%v", in.OrderSn, err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
