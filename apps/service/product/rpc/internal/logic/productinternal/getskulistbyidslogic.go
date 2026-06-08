package productinternallogic

import (
	"context"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSkuListByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSkuListByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSkuListByIdsLogic {
	return &GetSkuListByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSkuListByIds 批量获取Sku
func (l *GetSkuListByIdsLogic) GetSkuListByIds(in *product.SkuIdsReq) (*product.SkuListResp, error) {
	// 这里遵循部分成功原则
	list, err := l.svcCtx.ProductSkuModel.FindByIds(l.ctx, in.SkuIds)
	if err != nil {
		l.Logger.Errorf("批量获取SKU失败,error=%v", err)
		return nil, err
	}
	skuList := make([]*product.SkuItem, 0, len(*list))
	for _, sku := range *list {
		var specs []*product.SpecItem
		err := json.Unmarshal([]byte(sku.SpecData), &specs)
		if err != nil {
			l.Logger.Errorf("解析SKU数据失败,error=%v", err)
			return nil, err
		}
		var images []string
		if sku.Images.Valid {
			err := json.Unmarshal([]byte(sku.Images.String), &images)
			if err != nil {
				l.Logger.Errorf("解析SKU图片数据失败,error=%v", err)
				return nil, err
			}
		}
		skuList = append(skuList, &product.SkuItem{
			Id:          sku.Id,
			SpuId:       sku.SpuId,
			SpuName:     sku.SpuName,
			SkuCode:     sku.SkuCode,
			Price:       sku.Price,
			MarketPrice: sku.MarketPrice,
			CostPrice:   sku.CostPrice,
			Stock:       sku.Stock,
			LockedStock: sku.LockedStock,
			Version:     uint32(sku.Version),
			Specs:       specs,
			Images:      images,
			Weight:      sku.Weight,
			Status:      sku.Status,
			SpuStatus:   sku.SpuStatus,
		})
	}
	l.Logger.Info("批量获取SKU成功")
	return &product.SkuListResp{
		SkuItems: skuList,
	}, nil
}
