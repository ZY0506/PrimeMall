// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package list

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSkuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewBatchSkuLogic 批量获取SKU信息（价格、库存）
func NewBatchSkuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSkuLogic {
	return &BatchSkuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchSkuLogic) BatchSku(req *types.SkuIdsReq) (resp *types.BatchSkuResp, err error) {
	// 校验参数
	if len(req.SkuIds) == 0 {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "sku_ids 不能为空")
	}
	if len(req.SkuIds) > 100 {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "sku_ids 最多支持 100 个")
	}

	// 调用rpc
	skus, err := l.svcCtx.ProductRpc.BatchGetSkus(l.ctx, &product.SkuIdsReq{
		SkuIds: req.SkuIds,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc批量获取SKU信息失败，error=%v", err)
		return nil, err
	}

	list := make([]types.SkuSimple, 0)
	if skus.List == nil || len(skus.List) == 0 {
		return &types.BatchSkuResp{
			List: list,
		}, err
	}
	for _, sku := range skus.List {
		list = append(list, types.SkuSimple{
			Id:    sku.Id,
			Price: sku.Price,
			Stock: sku.Stock,
		})
	}

	return &types.BatchSkuResp{
		List: list,
	}, err
}
