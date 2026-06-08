package productinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetStockLogic {
	return &BatchGetStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BatchGetStock 批量查询库存（订单服务校验）
func (l *BatchGetStockLogic) BatchGetStock(in *product.SkuIdsReq) (*product.BatchStockResp, error) {
	// 这里遵循部分成功原则
	if len(in.SkuIds) == 0 {
		return &product.BatchStockResp{}, nil
	}
	skus, err := l.svcCtx.ProductSkuModel.FindByIds(l.ctx, in.SkuIds)
	if err != nil {
		l.Logger.Errorf("获取SKU列表失败，error=%v", err)
		return nil, err
	}
	// 返回响应
	stocks := make(map[uint64]int64)
	for _, sku := range *skus {
		stocks[sku.Id] = sku.Stock
	}
	return &product.BatchStockResp{
		Stocks: stocks,
	}, nil
}
