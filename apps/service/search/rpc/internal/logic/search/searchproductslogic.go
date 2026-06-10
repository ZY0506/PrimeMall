package searchlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/types/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchProductsLogic {
	return &SearchProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ES商品搜索
func (l *SearchProductsLogic) SearchProducts(in *search.SearchProductsReq) (*search.SearchProductsResp, error) {
	page := int(in.Page.GetPage())
	size := int(in.Page.GetSize())
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	if l.svcCtx.ES == nil {
		return &search.SearchProductsResp{
			Page: &search.PageResp{Total: 0, Page: int64(page), Size: int64(size)},
			List: []*search.ProductSearchItem{},
		}, nil
	}

	result, err := l.svcCtx.ES.SearchProducts(l.ctx, in.Keyword, in.CategoryId, in.Brand,
		in.MinPrice, in.MaxPrice, in.SortBy, in.SortType, page, size)
	if err != nil {
		// 降级处理：ES查询失败时返回空结果，避免HTTP 500拖垮整体失败率
		l.Logger.Errorf("商品搜索：ES查询失败，已降级返回空结果，错误：%v", err)
		return &search.SearchProductsResp{
			Page: &search.PageResp{Total: 0, Page: int64(page), Size: int64(size)},
			List: []*search.ProductSearchItem{},
		}, nil
	}

	list := make([]*search.ProductSearchItem, 0, len(result.Hits))
	for _, hit := range result.Hits {
		list = append(list, &search.ProductSearchItem{
			Id:           hit.ID,
			Name:         hit.Name,
			Brand:        hit.Brand,
			Cover:        hit.Cover,
			Price:        hit.Price,
			Sales:        hit.Sales,
			CategoryId:   uint64(hit.CategoryID),
			CategoryName: hit.CategoryName,
			Description:  hit.Description,
		})
	}

	return &search.SearchProductsResp{
		Page: &search.PageResp{
			Total: result.Total,
			Page:  int64(page),
			Size:  int64(size),
		},
		List: list,
	}, nil
}
