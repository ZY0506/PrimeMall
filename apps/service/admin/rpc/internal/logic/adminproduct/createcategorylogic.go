package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	productadmin "github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/productadmin"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCategoryLogic) CreateCategory(in *admin.AdminCreateCategoryReq) (*admin.Empty, error) {
	_, err := l.svcCtx.ProductRpc.CreateCategory(l.ctx, &productadmin.CreateCategoryReq{
		ParentId: in.ParentId,
		Name:     in.Name,
		Icon:     in.Icon,
		Sort:     in.Sort,
	})
	if err != nil {
		l.Logger.Errorf("创建分类失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
