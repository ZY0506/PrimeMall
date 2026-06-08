package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFreightTemplateLogic {
	return &UpdateFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateFreightTemplateLogic) UpdateFreightTemplate(in *admin.AdminUpdateFreightTemplateReq) (*admin.Empty, error) {
	_, err := l.svcCtx.ProductRpc.UpdateFreightTemplate(l.ctx, &producttypes.UpdateFreightTemplateReq{
		Id:                    in.Id,
		Name:                  in.Name,
		Type:                  int64(in.CalcType),
		DefaultFee:            in.DefaultFee,
		DefaultQuantity:       in.DefaultQuantity,
		ExtraFee:              in.ExtraFee,
		FreeThresholdAmount:   in.FreeThresholdAmount,
		FreeThresholdQuantity: in.FreeThresholdQuantity,
		IsDefault:             in.IsDefault,
		Status:                int64(in.Status),
	})
	if err != nil {
		l.Logger.Errorf("更新运费模板失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
