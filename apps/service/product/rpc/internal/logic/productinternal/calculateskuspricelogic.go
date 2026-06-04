package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CalculateSkusPriceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCalculateSkusPriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CalculateSkusPriceLogic {
	return &CalculateSkusPriceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CalculateSkusPriceLogic) CalculateSkusPrice(in *product.SkuIdsReq) (*product.CalculatePriceResp, error) {
	var items []*product.SkuStockItem
	for _, skuId := range in.SkuIds {
		items = append(items, &product.SkuStockItem{SkuId: skuId, Quantity: 1})
	}

	ret, err := l.svcCtx.ProductSkuModel.CalculateSkusPrice(l.ctx, items)
	if err != nil {
		return &product.CalculatePriceResp{}, err
	}
	l.Logger.Info("计算商品价格成功")

	return &product.CalculatePriceResp{
		TotalPrice: ret,
	}, nil
}
