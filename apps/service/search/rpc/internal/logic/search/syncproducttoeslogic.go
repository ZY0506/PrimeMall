package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncProductToESLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncProductToESLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncProductToESLogic {
	return &SyncProductToESLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步商品到ES
func (l *SyncProductToESLogic) SyncProductToES(in *search.SyncProductReq) (*search.Empty, error) {
	if l.svcCtx.ES == nil {
		return &search.Empty{}, nil
	}

	_ = l.svcCtx.ES.EnsureIndex(l.ctx)

	// 通过商品RPC获取完整的商品信息
	productDetail, err := l.svcCtx.ProductRpc.ProductDetail(l.ctx, &product.IdReq{Id: in.ProductId})
	if err != nil {
		l.Logger.Errorf("同步商品到ES：获取商品详情失败，productId=%d, error=%v", in.ProductId, err)
		return nil, err
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
	}
	// 如果有CategoryName，尝试从Attr获取或留空（搜索结果中会显示）
	_ = productDetail.Content // 备用

	// 构建搜索建议
	suggest := []string{productDetail.Name}
	if productDetail.Brand != "" {
		suggest = append(suggest, productDetail.Brand)
	}
	doc.Suggest = suggest

	if err := l.svcCtx.ES.IndexProduct(l.ctx, doc); err != nil {
		l.Logger.Errorf("同步商品到ES：索引失败，productId=%d, error=%v", in.ProductId, err)
		return nil, err
	}

	l.Logger.Infof("同步商品到ES成功，productId=%d, name=%s", in.ProductId, productDetail.Name)
	return &search.Empty{}, nil
}
