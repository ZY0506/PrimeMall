// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package category

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCategoryTreeLogic 获取商品全部分类树
func NewCategoryTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryTreeLogic {
	return &CategoryTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CategoryTreeLogic) CategoryTree() (resp *types.CategoryTreeResp, err error) {
	// 调用rpc，获取所有分类
	ret, err := l.svcCtx.ProductRpc.CategoryTree(l.ctx, &product.Empty{})
	if err != nil {
		l.Logger.Errorf("调用rpc查询分类出错，error=%v", err)
		return nil, err
	}

	list := make([]types.Category, 0)
	// 递归遍历赋值
	var assignment func(l []*product.Category) []types.Category
	assignment = func(l []*product.Category) []types.Category {
		var res []types.Category
		for i := range l {
			category := types.Category{
				Id:       l[i].Id,
				ParentId: l[i].ParentId,
				Name:     l[i].Name,
				Icon:     l[i].Icon,
				Level:    l[i].Level,
			}
			if len(l[i].Children) > 0 {
				category.Children = assignment(l[i].Children)
			}
			res = append(res, category)
		}
		return res
	}

	list = assignment(ret.List)

	return &types.CategoryTreeResp{
		List: list,
	}, nil
}
