package productlogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/go-redis/redis/v8"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListProducts 商品
func (l *ListProductsLogic) ListProducts(in *product.ProductListReq) (*product.ProductListResp, error) {
	// ===================== 缓存策略：仅缓存无筛选的简单列表 =====================
	isSimpleList := in.Keyword == "" && in.Brand == "" && in.MinPrice == 0 && in.MaxPrice == 0 &&
		len(in.Attrs) == 0 && in.SortBy == "" && !in.IsNew

	var listCacheKey string
	if isSimpleList {
		listCacheKey = fmt.Sprintf("%s%d:%d:%d", constants.ProductListKey, in.CategoryId, in.Page.Page, in.Page.Size)
		cached, err := l.svcCtx.Client.Get(l.ctx, listCacheKey).Result()
		if err == nil && cached != "" {
			var cachedResp product.ProductListResp
			if json.Unmarshal([]byte(cached), &cachedResp) == nil {
				l.Logger.Infof("商品列表缓存命中，categoryId=%d page=%d", in.CategoryId, in.Page.Page)
				return &cachedResp, nil
			}
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			l.Logger.Errorf("查询商品列表缓存失败，error=%v", err)
		}
	}

	// 要检验id
	category, err := l.svcCtx.CategoryModel.FindOne(l.ctx, in.CategoryId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("分类不存在,id=%d", in.CategoryId)
			return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "分类不存在")
		}
		l.Logger.Errorf("查询分类id出错，error=%v", err)
		return nil, err
	}
	if category.DeleteAt.Valid || category.Status != 1 {
		l.Logger.Errorf("分类不可用,id=%d", in.CategoryId)
		return nil, errorx.NewBizError(response.ErrCodeCategoryDisabled, "分类不可用")
	}

	var attrs []model.AttrFilter
	for _, attr := range in.Attrs {
		vals := make([]string, len(attr.Values))
		copy(vals, attr.Values)
		attrs = append(attrs, model.AttrFilter{
			Name:   attr.Name,
			Values: vals,
		})
	}
	// 构造筛选条件
	filters := &model.ListFilters{
		CategoryId: in.CategoryId,
		Keyword:    in.Keyword,
		Brand:      in.Brand,
		MinPrice:   in.MinPrice,
		MaxPrice:   in.MaxPrice,
		Attrs:      attrs,
		IsNew:      in.IsNew,
		SortBy:     in.SortBy,
		SortType:   in.SortType,
		Page:       in.Page.Page,
		PageSize:   in.Page.Size,
	}

	// 查询数据库
	ret, total, err := l.svcCtx.ProductSpuModel.FindListByFilter(l.ctx, filters)
	if err != nil {
		l.Logger.Errorf("查询商品出错，error=%v", err)
		return nil, err
	}
	list := make([]*product.ProductItem, 0)
	for _, item := range ret {
		list = append(list, &product.ProductItem{
			Id:        item.Id,
			Name:      item.Name,
			Brand:     item.Brand,
			Desc:      item.Desc,
			Cover:     item.Cover,
			Price:     item.Price,
			Sales:     item.Sales,
			ShowSales: item.ShowSales,
		})
	}
	resp := &product.ProductListResp{
		List: list,
		Page: &product.PageResp{
			Total: total,
			Page:  in.Page.Page,
			Size:  in.Page.Size,
		},
	}

	// ===================== 写入缓存 =====================
	if isSimpleList {
		if jsonBytes, marshalErr := json.Marshal(resp); marshalErr == nil {
			l.svcCtx.Client.Set(l.ctx, listCacheKey, string(jsonBytes), constants.ProductListTTL)
		}
	}

	l.Logger.Info("查询商品列表成功")
	return resp, nil
}
