package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func parseTime(s string) *timestamppb.Timestamp {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

type OrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *OrderListLogic) OrderList(req *types.AdminOrderListReq) (resp *types.AdminOrderListResp, err error) {
	rpcReq := &admin.AdminListOrdersReq{
		Page: req.Page, PageSize: req.Size,
		Status: admin.AdminOrderStatus(req.Status), OrderSn: req.OrderSn,
	}
	if req.StartTime != "" {
		rpcReq.StartTime = parseTime(req.StartTime)
	}
	if req.EndTime != "" {
		rpcReq.EndTime = parseTime(req.EndTime)
	}
	rpcResp, err := l.svcCtx.AdminOrderRpc.ListOrders(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	list := make([]types.AdminOrderSnapshot, 0, len(rpcResp.List))
	for _, o := range rpcResp.List {
		createdAt := ""
		if o.CreatedAt != nil {
			createdAt = o.CreatedAt.AsTime().Format(time.RFC3339)
		}
		list = append(list, types.AdminOrderSnapshot{
			OrderSn: o.OrderSn, UserId: o.UserId, UserPhone: o.UserPhone,
			PayAmount: o.PayAmount, Status: int64(o.Status),
			StatusDesc: o.StatusDesc, CreatedAt: createdAt,
		})
	}
	return &types.AdminOrderListResp{Total: rpcResp.Total, List: list}, nil
}
