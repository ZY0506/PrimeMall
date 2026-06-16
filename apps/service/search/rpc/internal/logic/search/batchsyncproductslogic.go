package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchSyncProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchSyncProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchSyncProductsLogic {
	return &BatchSyncProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量同步商品到ES
func (l *BatchSyncProductsLogic) BatchSyncProducts(in *search.BatchSyncProductsReq) (*search.Empty, error) {
	if l.svcCtx.ES == nil {
		return &search.Empty{}, nil
	}

	_ = l.svcCtx.ES.EnsureIndex(l.ctx)

	for _, pid := range in.ProductIds {
		// 通过商品RPC获取完整的商品信息
		productDetail, err := l.svcCtx.ProductRpc.ProductDetail(l.ctx, &product.IdReq{Id: pid})
		if err != nil {
			l.Logger.Errorf("批量同步商品：获取商品详情失败，productId=%d, error=%v", pid, err)
			continue
		}

		suggest := []string{productDetail.Name}
		if productDetail.Brand != "" {
			suggest = append(suggest, productDetail.Brand)
		}

		doc := &svc.ProductDoc{
			ID:           productDetail.Id,
			Name:         productDetail.Name,
			Brand:        productDetail.Brand,
			Description:  productDetail.Description,
			Cover:        productDetail.Cover,
			Price:        productDetail.Price,
			Sales:        productDetail.Sales,
			CategoryID:   int64(productDetail.CategoryId),
			CategoryName: "",
			Suggest:      suggest,
		}
		if err := l.svcCtx.ES.IndexProduct(l.ctx, doc); err != nil {
			l.Logger.Errorf("批量同步商品：索引商品%d失败，错误：%v", pid, err)
			continue
		}
		l.Logger.Infof("批量同步商品到ES成功，productId=%d, name=%s", pid, productDetail.Name)
	}
	return &search.Empty{}, nil
}
