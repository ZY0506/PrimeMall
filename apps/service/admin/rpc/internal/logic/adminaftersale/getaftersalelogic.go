package adminaftersalelogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	ordertypes "github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAfterSaleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAfterSaleLogic {
	return &GetAfterSaleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAfterSaleLogic) GetAfterSale(in *admin.AdminGetAfterSaleReq) (*admin.AdminAfterSaleDetailResp, error) {
	resp, err := l.svcCtx.OrderRpc.GetAfterSale(l.ctx, &ordertypes.AdminGetAfterSaleRequest{
		AfterSaleId: in.AfterSaleId,
	})
	if err != nil {
		l.Logger.Errorf("查询售后详情失败,after_sale_id=%d,error=%v", in.AfterSaleId, err)
		return nil, err
	}

	var productItem *admin.AdminOrderProductItem
	if resp.ProductItem != nil {
		productItem = &admin.AdminOrderProductItem{
			SkuId:       resp.ProductItem.SkuId,
			SpuId:       resp.ProductItem.SpuId,
			ProductName: resp.ProductItem.ProductName,
			SkuName:     resp.ProductItem.SkuName,
			Pic:         resp.ProductItem.Pic,
			Price:       resp.ProductItem.Price,
			Quantity:    resp.ProductItem.Quantity,
			TotalAmount: resp.ProductItem.TotalAmount,
		}
	}

	return &admin.AdminAfterSaleDetailResp{
		AfterSaleId:        resp.AfterSaleId,
		OrderSn:            resp.OrderSn,
		Type:               admin.AdminAfterSaleType(resp.Type),
		Status:             admin.AdminAfterSaleStatus(resp.Status),
		StatusDesc:         resp.StatusDesc,
		ApplyAmount:        resp.ApplyAmount,
		Reason:             resp.Reason,
		Images:             resp.Images,
		AuditRemark:        resp.AuditRemark,
		CreatedAt:          resp.CreatedAt,
		AuditTime:          resp.AuditTime,
		ReturnTrackingSn:   resp.ReturnTrackingSn,
		ReturnTrackingCorp: resp.ReturnTrackingCorp,
		RefundAmount:       resp.RefundAmount,
		RefundTime:         resp.RefundTime,
		ProductItem:        productItem,
	}, nil
}
