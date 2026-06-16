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

type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 商品列表与筛选
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListLogic) List(req *types.ProductListReq) (resp *types.ProductListResp, err error) {
	// 校验排序参数
	if req.SortBy != "" {
		switch req.SortBy {
		case "price", "sales", "created_at":
			// 有效的排序字段，校验排序方向
			if req.SortType != "" && req.SortType != "asc" && req.SortType != "desc" {
				return nil, response.NewBizError(response.ErrCodeInvalidParam, "无效的排序方向，仅支持 asc/desc")
			}
		default:
			return nil, response.NewBizError(response.ErrCodeInvalidParam, "无效的排序字段，仅支持 price/sales/created_at")
		}
	} else if req.SortType != "" {
		return nil, response.NewBizError(response.ErrCodeInvalidParam, "指定排序方向时必须同时指定排序字段")
	}

	var attrs []*product.AttrFilter
	for _, attr := range req.Attrs {
		attrs = append(attrs, &product.AttrFilter{
			Name:   attr.Name,
			Values: attr.Values,
		})
	}

	// 调用rpc获取商品列表
	ret, err := l.svcCtx.ProductRpc.ListProducts(l.ctx, &product.ProductListReq{
		CategoryId: req.CategoryId,
		Brand:      req.Brand,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Keyword:    req.Keyword,
		Attrs:      attrs,
		IsNew:      req.IsNew,
		SortBy:     req.SortBy,
		SortType:   req.SortType,
		Page: &product.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
	})
	if err != nil {
		l.Logger.Errorf("调用rpc获取商品列表失败，err=%v", err)
		return
	}
	var list []types.ProductItem
	for _, item := range ret.List {
		list = append(list, types.ProductItem{
			Id:         item.Id,
			Name:       item.Name,
			Brand:      item.Brand,
			DefaultPic: item.Cover,
			Price:      item.Price,
			Sales:      item.Sales,
			ShowSales:  item.ShowSales,
		})
	}
	return &types.ProductListResp{
		List:  list,
		Total: ret.Page.Total,
	}, nil
}
