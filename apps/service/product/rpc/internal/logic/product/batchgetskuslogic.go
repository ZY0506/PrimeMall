package productlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetSkusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetSkusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetSkusLogic {
	return &BatchGetSkusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BatchGetSkus SKU
func (l *BatchGetSkusLogic) BatchGetSkus(in *product.SkuIdsReq) (*product.BatchSkuResp, error) {
	// 这里遵循部分成功原则
	list, err := l.svcCtx.ProductSkuModel.FindByIds(l.ctx, in.SkuIds)
	if err != nil {
		l.Logger.Errorf("批量获取SKU失败,error=%v", err)
		return nil, err
	}
	// 转换为SKU简单信息
	skuSimple := make([]*product.SkuSimple, 0)
	for _, item := range *list {
		skuSimple = append(skuSimple, &product.SkuSimple{
			Id:    item.Id,
			Price: item.Price,
			Stock: item.Stock,
		})
	}

	return &product.BatchSkuResp{
		List: skuSimple,
	}, nil
}
