package adminaftersalelogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAfterSalesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAfterSalesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAfterSalesLogic {
	return &ListAfterSalesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAfterSalesLogic) ListAfterSales(in *admin.AdminListAfterSalesReq) (*admin.AdminListAfterSalesResp, error) {
	resp, err := l.svcCtx.OrderRpc.ListAfterSales(l.ctx, &ordertypes.AdminListAfterSalesRequest{
		Page:     in.Page,
		PageSize: in.PageSize,
		Status:   ordertypes.AfterSaleStatus(in.Status),
	})
	if err != nil {
		l.Logger.Errorf("查询售后列表失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminAfterSaleItem
	for _, item := range resp.List {
		list = append(list, &admin.AdminAfterSaleItem{
			AfterSaleId:  item.AfterSaleId,
			OrderSn:      item.OrderSn,
			Type:         admin.AdminAfterSaleType(item.Type),
			Status:       admin.AdminAfterSaleStatus(item.Status),
			StatusDesc:   item.StatusDesc,
			RefundAmount: item.RefundAmount,
			SpuName:      item.SpuName,
			SkuName:      item.SkuName,
			Pic:          item.Pic,
			CreatedAt:    item.CreatedAt,
		})
	}

	return &admin.AdminListAfterSalesResp{
		Total: resp.Total,
		List:  list,
	}, nil
}
