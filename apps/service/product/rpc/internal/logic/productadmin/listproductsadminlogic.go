package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsAdminLogic {
	return &ListProductsAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductsAdminLogic) ListProductsAdmin(in *product.ListProductsAdminReq) (*product.ListProductsAdminResp, error) {
	if in.Page == nil {
		in.Page = &product.PageReq{Page: 1, Size: 10}
	}
	if in.Page.Page <= 0 {
		in.Page.Page = 1
	}
	if in.Page.Size <= 0 {
		in.Page.Size = 10
	}

	filters := &model.ListFilters{
		CategoryId: in.CategoryId,
		Brand:      in.Brand,
		MinPrice:   in.MinPrice,
		MaxPrice:   in.MaxPrice,
		Keyword:    in.Keyword,
		SortBy:     in.SortBy,
		SortType:   in.SortType,
		Page:       in.Page.Page,
		PageSize:   in.Page.Size,
	}

	items, total, err := l.svcCtx.ProductSpuModel.FindListByFilter(l.ctx, filters)
	if err != nil {
		l.Logger.Errorf("查询商品列表失败, err=%v", err)
		return nil, err
	}

	var list []*product.ProductAdminItem
	for _, item := range items {
		status := product.ProductStatus_PRODUCT_STATUS_OFF_SALE
		if item.Price > 0 {
			status = product.ProductStatus_PRODUCT_STATUS_ON_SALE
		}
		list = append(list, &product.ProductAdminItem{
			Id:         item.Id,
			Name:       item.Name,
			Brand:      item.Brand,
			Cover:      item.Cover,
			Price:      item.Price,
			SalesCount: item.Sales,
			Status:     status,
		})
	}

	l.Logger.Infof("查询商品列表成功, total=%d", total)
	return &product.ListProductsAdminResp{
		Page: &product.PageResp{
			Total: total,
			Page:  in.Page.Page,
			Size:  in.Page.Size,
		},
		List: list,
	}, nil
}
