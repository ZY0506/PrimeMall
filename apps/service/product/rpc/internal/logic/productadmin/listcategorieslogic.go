package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListCategoriesLogic) ListCategories(in *product.ListCategoriesReq) (*product.ListCategoriesResp, error) {
	if in.Page == nil {
		in.Page = &product.PageReq{Page: 1, Size: 10}
	}
	if in.Page.Page <= 0 {
		in.Page.Page = 1
	}
	if in.Page.Size <= 0 {
		in.Page.Size = 10
	}

	categories, err := l.svcCtx.CategoryModel.FindALL(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询分类列表失败, err=%v", err)
		return nil, err
	}

	var list []*product.Category
	if categories != nil {
		for _, cat := range *categories {
			if in.ParentId > 0 && cat.ParentId != in.ParentId {
				continue
			}
			if in.ParentId == 0 && cat.ParentId != 0 {
				continue
			}
			list = append(list, &product.Category{
				Id:       cat.Id,
				ParentId: cat.ParentId,
				Name:     cat.Name,
				Icon:     cat.Icon,
				Sort:     cat.Sort,
				Level:    cat.Level,
				Status:   cat.Status,
			})
		}
	}

	l.Logger.Infof("查询分类列表成功, total=%d", len(list))
	return &product.ListCategoriesResp{
		Page: &product.PageResp{
			Total: int64(len(list)),
			Page:  in.Page.Page,
			Size:  in.Page.Size,
		},
		List: list,
	}, nil
}
