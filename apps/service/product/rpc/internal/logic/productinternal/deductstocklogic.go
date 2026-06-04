package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductStockLogic {
	return &DeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeductStock 扣减库存（订单支付时）
func (l *DeductStockLogic) DeductStock(in *product.UpdateStockReq) (*product.StockChangeResp, error) {
	if in.OrderSn == "" {
		return &product.StockChangeResp{}, errorx.NewBizError(response.ErrCodeOrderNotFound, "订单号不能为空")
	}
	for _, item := range in.Items {
		if item.Quantity <= 0 {
			return &product.StockChangeResp{}, errorx.NewBizError(response.ErrCodeInvalidQuantity, "数量必须大于0")
		}
	}

	skuIds := make([]uint64, len(in.Items))
	for i, item := range in.Items {
		skuIds[i] = item.SkuId
	}
	if !acquireStockLocks(l.ctx, l.svcCtx.Client, skuIds) {
		return &product.StockChangeResp{}, errorx.NewBizError(response.ErrCodeTooFrequent, "系统繁忙，请稍后重试")
	}
	defer releaseStockLocksByIds(l.ctx, l.svcCtx.Client, skuIds)

	ret, err := l.svcCtx.ProductSkuModel.DeductStock(l.ctx, in.Items, in.OrderSn)
	if err != nil {
		return &product.StockChangeResp{}, err
	}
	l.Logger.Infof("订单：%s 扣减库存成功", in.OrderSn)

	return &product.StockChangeResp{
		Success: true,
		Results: *ret,
	}, nil
}
