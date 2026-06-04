package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevertDeductStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevertDeductStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevertDeductStockLogic {
	return &RevertDeductStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RevertDeductStock 恢复库存(下单事务失败时回滚)
func (l *RevertDeductStockLogic) RevertDeductStock(in *product.UpdateStockReq) (*product.StockChangeResp, error) {
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

	ret, err := l.svcCtx.ProductSkuModel.RevertDeduct(l.ctx, in.Items, in.OrderSn)
	if err != nil {
		l.Logger.Errorf("订单：%s 恢复库存失败：%v", in.OrderSn, err)
		return &product.StockChangeResp{}, err
	}
	l.Logger.Infof("订单：%s 恢复库存成功", in.OrderSn)
	return &product.StockChangeResp{
		Success: true,
		Results: *ret,
	}, nil
}
