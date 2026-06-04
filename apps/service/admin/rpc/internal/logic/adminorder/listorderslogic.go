package adminorderlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *admin.AdminListOrdersReq) (*admin.AdminListOrdersResp, error) {
	resp, err := l.svcCtx.OrderRpc.ListOrders(l.ctx, &ordertypes.AdminListOrdersRequest{
		Page:      in.Page,
		PageSize:  in.PageSize,
		Status:    ordertypes.OrderStatus(in.Status),
		OrderSn:   in.OrderSn,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
	})
	if err != nil {
		l.Logger.Errorf("查询订单列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminOrderItem
	for _, item := range resp.List {
		list = append(list, &admin.AdminOrderItem{
			OrderSn:    item.OrderSn,
			UserId:     item.UserId,
			UserPhone:  item.UserPhone,
			PayAmount:  item.PayAmount,
			Status:     admin.AdminOrderStatus(item.Status),
			StatusDesc: item.StatusDesc,
			CreatedAt:  item.CreatedAt,
		})
	}

	return &admin.AdminListOrdersResp{
		Total: resp.Total,
		List:  list,
	}, nil
}
