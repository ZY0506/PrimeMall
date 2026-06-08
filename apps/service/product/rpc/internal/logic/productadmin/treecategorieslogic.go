package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type TreeCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTreeCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TreeCategoriesLogic {
	return &TreeCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TreeCategoriesLogic) TreeCategories(in *product.Empty) (*product.CategoryTreeResp, error) {
	allCategories, err := l.svcCtx.CategoryModel.FindALL(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询所有分类失败, err=%v", err)
		return nil, err
	}

	tree := buildCategoryTree(*allCategories, 0)
	l.Logger.Infof("获取分类树成功, total=%d", len(*allCategories))
	return &product.CategoryTreeResp{List: tree}, nil
}

func buildCategoryTree(categories []model.Category, parentId uint64) []*product.Category {
	var result []*product.Category
	for _, cat := range categories {
		if cat.ParentId == parentId {
			item := &product.Category{
				Id:       cat.Id,
				ParentId: cat.ParentId,
				Name:     cat.Name,
				Icon:     cat.Icon,
				Sort:     cat.Sort,
				Level:    cat.Level,
				Status:   cat.Status,
			}
			item.Children = buildCategoryTree(categories, cat.Id)
			result = append(result, item)
		}
	}
	return result
}
