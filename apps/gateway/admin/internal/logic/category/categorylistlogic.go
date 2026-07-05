// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package category

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCategoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CategoryListLogic {
	return &CategoryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CategoryListLogic) CategoryList() (resp *types.CategoryListResp, err error) {
	rpcResp, err := l.svcCtx.AdminProductRpc.CategoryTree(l.ctx, &admin.Empty{})
	if err != nil {
		return nil, err
	}
	list := make([]types.AdminCategoryItem, 0, len(rpcResp.List))
	for _, c := range rpcResp.List {
		list = append(list, types.AdminCategoryItem{
			Id: c.Id, ParentId: c.ParentId, Name: c.Name,
			Icon: c.Icon, Level: c.Level, Sort: c.Sort,
		})
	}
	return &types.CategoryListResp{List: list}, nil
}
