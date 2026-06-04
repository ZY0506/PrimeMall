package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFreightTemplatesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFreightTemplatesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFreightTemplatesLogic {
	return &ListFreightTemplatesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFreightTemplatesLogic) ListFreightTemplates(in *product.PageReq) (*product.ListFreightTemplatesResp, error) {
	if in == nil {
		in = &product.PageReq{Page: 1, Size: 10}
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Size <= 0 {
		in.Size = 10
	}

	templates, total, err := l.svcCtx.FreightTemplateModel.FindList(l.ctx, in.Page, in.Size)
	if err != nil {
		l.Logger.Errorf("查询运费模板列表失败, err=%v", err)
		return nil, err
	}

	var list []*product.FreightTemplate
	for _, template := range templates {
		list = append(list, &product.FreightTemplate{
			Id:                    template.Id,
			Name:                  template.Name,
			Type:                  template.Type,
			DefaultFee:            template.DefaultFee,
			DefaultQuantity:       template.DefaultQuantity,
			ExtraFee:              template.ExtraFee,
			FreeThresholdAmount:   template.FreeThresholdAmount,
			FreeThresholdQuantity: template.FreeThresholdQuantity,
			IsDefault:             template.IsDefault,
			Status:                template.Status,
			CreatedAt:             template.CreatedAt.Unix(),
			UpdatedAt:             template.UpdatedAt.Unix(),
		})
	}

	l.Logger.Infof("查询运费模板列表成功, total=%d", total)
	return &product.ListFreightTemplatesResp{
		Page: &product.PageResp{
			Total: total,
			Page:  in.Page,
			Size:  in.Size,
		},
		List: list,
	}, nil
}
