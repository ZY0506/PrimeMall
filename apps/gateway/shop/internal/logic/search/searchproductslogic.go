package search

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchProductsLogic {
	return &SearchProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchProductsLogic) SearchProducts(req *types.SearchProductsReq) (resp *types.SearchProductsResp, err error) {
	rpcResp, err := l.svcCtx.SearchRpc.SearchProducts(l.ctx, &search.SearchProductsReq{
		Keyword:    req.Keyword,
		CategoryId: req.CategoryId,
		Brand:      req.Brand,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		SortBy:     req.SortBy,
		SortType:   req.SortType,
		Page:       &search.PageReq{Page: req.Page, Size: req.Size},
	})
	if err != nil {
		l.Logger.Errorf("SearchProducts RPC error: %v", err)
		return nil, err
	}

	list := make([]types.SearchProductItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.SearchProductItem{
			Id:           item.Id,
			Name:         item.Name,
			Brand:        item.Brand,
			Cover:        item.Cover,
			Price:        item.Price,
			Sales:        item.Sales,
			CategoryId:   item.CategoryId,
			CategoryName: item.CategoryName,
			Description:  item.Description,
		})
	}

	return &types.SearchProductsResp{
		Total: rpcResp.Page.Total,
		List:  list,
	}, nil
}
