package productlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

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
	list, err := l.svcCtx.ProductSkuModel.FindByIds(l.ctx, in.SkuIds)
	if err != nil {
		l.Logger.Errorf("批量获取SKU失败,error=%v", err)
		return nil, err
	}

	// 检查是否有未找到的 SKU（存在不存在的 SKU ID 时返回错误）
	if len(*list) < len(in.SkuIds) {
		l.Logger.Errorf("批量获取SKU失败：部分SKU不存在,ids=%v", in.SkuIds)
		return nil, errorx.NewBizError(response.ErrCodeSkuNotFound, "部分SKU不存在")
	}

	// 转换为SKU简单信息
	skuSimple := make([]*product.SkuSimple, 0, len(*list))
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
