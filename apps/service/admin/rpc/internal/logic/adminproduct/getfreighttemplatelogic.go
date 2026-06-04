package adminproductlogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	producttypes "github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetFreightTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFreightTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFreightTemplateLogic {
	return &GetFreightTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFreightTemplateLogic) GetFreightTemplate(in *admin.IdReq) (*admin.AdminFreightTemplateInfo, error) {
	resp, err := l.svcCtx.ProductRpc.GetFreightTemplate(l.ctx, &producttypes.IdReq{Id: in.Id})
	if err != nil {
		l.Logger.Errorf("获取运费模板失败,error=%v", err)
		return nil, err
	}

	return &admin.AdminFreightTemplateInfo{
		Id:                    resp.Id,
		Name:                  resp.Name,
		CalcType:              admin.FreightCalcType(resp.Type),
		DefaultFee:            resp.DefaultFee,
		DefaultQuantity:       resp.DefaultQuantity,
		ExtraFee:              resp.ExtraFee,
		FreeThresholdAmount:   resp.FreeThresholdAmount,
		FreeThresholdQuantity: resp.FreeThresholdQuantity,
		IsDefault:             resp.IsDefault,
		Status:                admin.SwitchStatus(resp.Status),
		CreatedAt:             timestamppb.New(time.Unix(resp.CreatedAt, 0)),
	}, nil
}
