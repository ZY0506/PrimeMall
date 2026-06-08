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

type CategoryInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCategoryInfoLogic 获取单个分类详情
func NewCategoryInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryInfoLogic {
	return &CategoryInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CategoryInfoLogic) CategoryInfo(req *types.IdPathReq) (resp *types.CategoryResp, err error) {
	// 调用rpc
	ret, err := l.svcCtx.ProductRpc.CategoryInfo(l.ctx, &product.IdReq{Id: req.Id})
	if err != nil {
		l.Logger.Errorf("调用rpc获取分类信息失败，error=%v", err)
		return nil, err
	}

	var convertCategory func(c *product.Category) types.Category
	convertCategory = func(c *product.Category) types.Category {
		// 转换当前节点
		res := types.Category{
			Id:       c.Id,
			ParentId: c.ParentId,
			Name:     c.Name,
			Icon:     c.Icon,
			Level:    c.Level,
		}

		// 2. 递归处理子节点
		if c.Children != nil && len(c.Children) > 0 {
			for _, child := range c.Children {
				res.Children = append(res.Children, convertCategory(child))
			}
		}
		return res
	}

	return &types.CategoryResp{
		Category: convertCategory(ret.Category),
	}, nil
}
