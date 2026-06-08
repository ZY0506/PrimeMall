// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package list

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewProductDetailLogic 商品详情
func NewProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductDetailLogic {
	return &ProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProductDetailLogic) ProductDetail(req *types.IdPathReq) (resp *types.ProductDetailResp, err error) {
	// 调用rpc
	ret, err := l.svcCtx.ProductRpc.ProductDetail(l.ctx, &product.IdReq{
		Id: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc获取商品详情失败，error=%v", err)
		return nil, err
	}

	// 构造返回响应
	skus := make([]types.SkuItem, 0)
	for _, item := range ret.Skus {
		specs := make([]types.SpecItem, 0)
		for _, item := range item.Specs {
			specs = append(specs, types.SpecItem{
				Name:   item.Key,
				Values: item.Value,
			})
		}
		skus = append(skus, types.SkuItem{
			Id:       item.Id,
			Code:     item.SkuCode,
			Pic:      item.Images,
			Price:    item.Price,
			SpecData: specs,
			Stock:    item.Stock,
			Weight:   item.Weight,
		})
	}
	attributes := make([]types.AttributeItem, 0)
	for _, item := range ret.Attributes {
		attributes = append(attributes, types.AttributeItem{
			Name:   item.Name,
			Values: item.Values,
		})
	}

	return &types.ProductDetailResp{
		Id:                ret.Id,
		CategoryId:        ret.CategoryId,
		Name:              ret.Name,
		Brand:             ret.Brand,
		Description:       ret.Description,
		Content:           ret.Content,
		DefaultPic:        ret.Cover,
		BannerPics:        ret.Images,
		VideoUrl:          ret.VideoUrl,
		FreightTemplateId: ret.FreightTemplateId,
		Price:             ret.Price, // 最低价
		Sales:             ret.Sales,
		ShowSales:         ret.ShowSales,
		Skus:              skus,
		Attributes:        attributes, // 聚合属性（用于前端渲染）
	}, nil
}
