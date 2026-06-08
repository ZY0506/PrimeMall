package adminproductlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFreightTemplateLogic {
	return &CreateFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 运费模板管理
func (l *CreateFreightTemplateLogic) CreateFreightTemplate(in *admin.AdminCreateFreightTemplateReq) (*admin.Empty, error) {
	_, err := l.svcCtx.ProductRpc.CreateFreightTemplate(l.ctx, &producttypes.CreateFreightTemplateReq{
		Name:                  in.Name,
		Type:                  int64(in.CalcType),
		DefaultFee:            in.DefaultFee,
		DefaultQuantity:       in.DefaultQuantity,
		ExtraFee:              in.ExtraFee,
		FreeThresholdAmount:   in.FreeThresholdAmount,
		FreeThresholdQuantity: in.FreeThresholdQuantity,
		IsDefault:             in.IsDefault,
	})
	if err != nil {
		l.Logger.Errorf("创建运费模板失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
