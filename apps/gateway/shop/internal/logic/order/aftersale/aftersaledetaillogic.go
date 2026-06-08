// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package aftersale

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"time"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AfterSaleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAfterSaleDetailLogic 查询售后详情
func NewAfterSaleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AfterSaleDetailLogic {
	return &AfterSaleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AfterSaleDetailLogic) AfterSaleDetail(req *types.IdPathReq) (resp *types.AfterSaleDetailResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	afterSaleDetailResp, err := l.svcCtx.OrderRpc.AfterSaleDetail(l.ctx, &order.AfterSaleDetailRequest{
		AfterSaleId: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("调用RPC查询售后详情失败，error=%v", err)
		return nil, err
	}

	return &types.AfterSaleDetailResp{
		AfterSaleId:        afterSaleDetailResp.AfterSaleId,
		OrderSn:            afterSaleDetailResp.OrderSn,
		AfterSaleType:      int64(afterSaleDetailResp.Type),
		Status:             int64(afterSaleDetailResp.Status),
		StatusDesc:         afterSaleDetailResp.StatusDesc,
		ApplyAmount:        afterSaleDetailResp.ApplyAmount,
		Reason:             afterSaleDetailResp.Reason,
		Images:             afterSaleDetailResp.Images,
		AuditRemark:        afterSaleDetailResp.AuditRemark,
		CreateTime:         afterSaleDetailResp.CreatedAt.AsTime().Format(time.RFC3339),
		AuditTime:          afterSaleDetailResp.AuditTime.AsTime().Format(time.RFC3339),
		ReturnTrackingSn:   afterSaleDetailResp.ReturnTrackingSn,
		ReturnTrackingCorp: afterSaleDetailResp.ReturnTrackingCorp,
		ReceiveTime:        afterSaleDetailResp.ReceiveTime.AsTime().Format(time.RFC3339),
		RefundAmount:       afterSaleDetailResp.RefundAmount,
		RefundTime:         afterSaleDetailResp.RefundTime.AsTime().Format(time.RFC3339),
		OrderItem: types.OrderItem{
			SkuId:       afterSaleDetailResp.ProductItem.SkuId,
			SpuId:       afterSaleDetailResp.ProductItem.SpuId,
			ProductName: afterSaleDetailResp.ProductItem.ProductName,
			SkuName:     afterSaleDetailResp.ProductItem.SkuName,
			Pic:         afterSaleDetailResp.ProductItem.Pic,
			Price:       afterSaleDetailResp.ProductItem.Price,
			Quantity:    afterSaleDetailResp.ProductItem.Quantity,
			TotalAmount: afterSaleDetailResp.ProductItem.TotalAmount,
		},
	}, nil
}
