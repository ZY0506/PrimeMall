package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	productadmin "github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productadmin"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCategoryLogic) UpdateCategory(in *admin.AdminUpdateCategoryReq) (*admin.Empty, error) {
	_, err := l.svcCtx.ProductRpc.UpdateCategory(l.ctx, &productadmin.UpdateCategoryReq{
		Id:     in.Id,
		Name:   in.Name,
		Icon:   in.Icon,
		Sort:   in.Sort,
		Status: int64(in.Status),
	})
	if err != nil {
		l.Logger.Errorf("更新分类失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
