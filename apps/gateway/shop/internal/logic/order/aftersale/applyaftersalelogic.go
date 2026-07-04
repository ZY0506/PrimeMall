// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package aftersale

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyAfterSaleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewApplyAfterSaleLogic 申请售后（幂等）
func NewApplyAfterSaleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAfterSaleLogic {
	return &ApplyAfterSaleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyAfterSaleLogic) ApplyAfterSale(req *types.AfterSaleReq) (resp *types.AfterSaleResp, err error) {
	// 注入userID
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}
	// 调用rpc
	_, err = l.svcCtx.OrderRpc.ApplyAfterSale(l.ctx, &order.AfterSaleRequest{
		OrderSn:        req.OrderSn,
		SkuId:          req.SkuId,
		Quantity:       req.Quantity,
		Type:           order.AfterSaleType(req.Type),
		Reason:         req.Reason,
		ApplyAmount:    req.ApplyAmount,
		Images:         req.Images,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc申请售后失败，error=%v", err)
		return nil, err
	}

	return &types.AfterSaleResp{
		Msg: "申请售后成功，等待管理员审核",
	}, nil
}
