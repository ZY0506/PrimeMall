package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductStatusLogic {
	return &UpdateProductStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductStatusLogic) UpdateProductStatus(in *admin.AdminUpdateProductStatusReq) (*admin.Empty, error) {
	// 通过更新商品接口更新状态
	_, err := l.svcCtx.ProductRpc.UpdateProduct(l.ctx, &producttypes.UpdateProductReq{
		Id:     in.Id,
		Status: producttypes.ProductStatus(in.Status),
	})
	if err != nil {
		l.Logger.Errorf("更新商品状态失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
