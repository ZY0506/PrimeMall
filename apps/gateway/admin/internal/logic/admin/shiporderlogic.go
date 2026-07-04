package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type ShipOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShipOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShipOrderLogic {
	return &ShipOrderLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *ShipOrderLogic) ShipOrder(req *types.ShipOrderReq) error {
	if req.OrderSn == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "订单号不能为空")
	}
	if req.DeliveryCorp == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "物流公司不能为空")
	}
	if req.DeliverySn == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "物流单号不能为空")
	}
	_, err := l.svcCtx.AdminOrderRpc.ShipOrder(l.ctx, &admin.AdminShipOrderReq{
		OrderSn: req.OrderSn, DeliveryCompany: req.DeliveryCorp,
		DeliverySn: req.DeliverySn,
	})
	if err != nil {
		l.Logger.Errorf("订单发货失败 orderSn=%s: %v", req.OrderSn, err)
		return err
	}
	l.Logger.Infof("订单发货成功 orderSn=%s company=%s", req.OrderSn, req.DeliveryCorp)
	return nil
}
