package productadminlogic

import (
	"context"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductAdminLogic {
	return &GetProductAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductAdminLogic) GetProductAdmin(in *product.IdReq) (*product.ProductDetailResp, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "商品ID不能为空")
	}

	spu, err := l.svcCtx.ProductSpuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewBizError(response.ErrCodeProductNotFound, "商品不存在")
		}
		l.Logger.Errorf("查询商品失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	var images []string
	if spu.SubPics.Valid && spu.SubPics.String != "" {
		json.Unmarshal([]byte(spu.SubPics.String), &images)
	}

	var description string
	if spu.Desc != "" {
		description = spu.Desc
	}

	var content string
	if spu.Content.Valid {
		content = spu.Content.String
	}

	skus, err := l.svcCtx.ProductSkuModel.FindListBySpuId(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("查询SKU列表失败, spuId=%d, err=%v", in.Id, err)
		return nil, err
	}

	var skuItems []*product.SkuItem
	for _, sku := range *skus {
		var specs []*product.SpecItem
		if sku.SpecData != "" {
			json.Unmarshal([]byte(sku.SpecData), &specs)
		}

		var skuImages []string
		if sku.Images.Valid && sku.Images.String != "" {
			json.Unmarshal([]byte(sku.Images.String), &skuImages)
		}

		skuItems = append(skuItems, &product.SkuItem{
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
			Images:      skuImages,
			Weight:      sku.Weight,
			Status:      sku.Status,
			SpuStatus:   spu.Status,
		})
	}

	minPrice := int64(0)
	for i, sku := range skuItems {
		if i == 0 || sku.Price < minPrice {
			minPrice = sku.Price
		}
	}

	resp := &product.ProductDetailResp{
		Id:                spu.Id,
		CategoryId:        spu.CategoryId,
		Name:              spu.Name,
		Brand:             spu.Brand,
		Description:       description,
		Content:           content,
		Cover:             spu.MainPic,
		Images:            images,
		VideoUrl:          spu.VideoUrl,
		Price:             minPrice,
		Sales:             spu.SalesCount + spu.VirtualSales,
		ShowSales:         spu.SalesCount + spu.VirtualSales,
		Status:            product.ProductStatus(spu.Status),
		FreightTemplateId: spu.FreightTemplateId,
		Skus:              skuItems,
		SalesCount:        spu.SalesCount,
		VirtualSales:      spu.VirtualSales,
	}

	l.Logger.Infof("获取商品详情成功, id=%d", in.Id)
	return resp, nil
}
