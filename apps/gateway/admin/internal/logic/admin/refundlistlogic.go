package admin

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefundListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefundListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundListLogic {
	return &RefundListLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *RefundListLogic) RefundList(req *types.AdminRefundListReq) (resp *types.AdminRefundListResp, err error) {
	rpcResp, err := l.svcCtx.AdminAfterSaleRpc.ListAfterSales(l.ctx, &admin.AdminListAfterSalesReq{
		Page: req.Page, PageSize: req.Size, Status: admin.AdminAfterSaleStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.AdminRefundItem, 0, len(rpcResp.List))
	for _, r := range rpcResp.List {
		createdAt := ""
		if r.CreatedAt != nil {
			createdAt = r.CreatedAt.AsTime().Format("2006-01-02 15:04:05")
		}
		list = append(list, types.AdminRefundItem{
			Id: r.AfterSaleId, OrderSn: r.OrderSn, RefundAmount: r.RefundAmount,
			Status: int64(r.Status), StatusDesc: r.StatusDesc, CreatedAt: createdAt,
		})
	}
	return &types.AdminRefundListResp{Total: rpcResp.Total, List: list}, nil
}
