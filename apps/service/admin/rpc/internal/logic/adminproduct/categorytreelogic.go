package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryTreeLogic {
	return &CategoryTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分类管理
func (l *CategoryTreeLogic) CategoryTree(in *admin.Empty) (*admin.AdminCategoryTreeResp, error) {
	resp, err := l.svcCtx.ProductRpc.TreeCategories(l.ctx, &producttypes.Empty{})
	if err != nil {
		l.Logger.Errorf("获取分类树失败,error=%v", err)
		return nil, err
	}

	var list []*admin.AdminCategoryItem
	for _, item := range resp.List {
		list = append(list, convertCategory(item))
	}

	return &admin.AdminCategoryTreeResp{
		List: list,
	}, nil
}

func convertCategory(c *producttypes.Category) *admin.AdminCategoryItem {
	if c == nil {
		return nil
	}
	item := &admin.AdminCategoryItem{
		Id:       c.Id,
		ParentId: c.ParentId,
		Name:     c.Name,
		Icon:     c.Icon,
		Sort:     c.Sort,
		Level:    c.Level,
		Status:   admin.SwitchStatus(c.Status),
	}
	for _, child := range c.Children {
		item.Children = append(item.Children, convertCategory(child))
	}
	return item
}
