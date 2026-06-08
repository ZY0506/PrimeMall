package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFreightTemplateLogic {
	return &DeleteFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteFreightTemplateLogic) DeleteFreightTemplate(in *admin.IdReq) (*admin.Empty, error) {
	_, err := l.svcCtx.ProductRpc.DeleteFreightTemplate(l.ctx, &producttypes.IdReq{Id: in.Id})
	if err != nil {
		l.Logger.Errorf("删除运费模板失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
