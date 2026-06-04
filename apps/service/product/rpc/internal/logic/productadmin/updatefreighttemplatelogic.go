package productadminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

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

func (l *UpdateFreightTemplateLogic) UpdateFreightTemplate(in *product.UpdateFreightTemplateReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "模板ID不能为空")
	}

	template, err := l.svcCtx.FreightTemplateModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewBizError(response.ErrCodeFreightTemplateNotFound, "运费模板不存在")
		}
		l.Logger.Errorf("查询运费模板失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	if in.Name != "" {
		template.Name = in.Name
	}
	if in.Type != 0 {
		if in.Type != 1 && in.Type != 2 {
			return nil, errorx.NewBizError(response.ErrCodeParamType, "计费方式只能是1或2")
		}
		template.Type = in.Type
	}
	if in.DefaultFee != 0 {
		template.DefaultFee = in.DefaultFee
	}
	if in.DefaultQuantity != 0 {
		template.DefaultQuantity = in.DefaultQuantity
	}
	if in.ExtraFee != 0 {
		template.ExtraFee = in.ExtraFee
	}
	if in.FreeThresholdAmount != 0 {
		template.FreeThresholdAmount = in.FreeThresholdAmount
	}
	if in.FreeThresholdQuantity != 0 {
		template.FreeThresholdQuantity = in.FreeThresholdQuantity
	}
	if in.IsDefault != 0 {
		if in.IsDefault == 1 {
			defaultTemplate, err := l.svcCtx.FreightTemplateModel.FindDefault(l.ctx)
			if err == nil && defaultTemplate != nil && defaultTemplate.Id != in.Id {
				defaultTemplate.IsDefault = 0
				err = l.svcCtx.FreightTemplateModel.Update(l.ctx, defaultTemplate)
				if err != nil {
					l.Logger.Errorf("更新默认模板失败, err=%v", err)
					return nil, err
				}
			}
		}
		template.IsDefault = in.IsDefault
	}
	if in.Status != 0 {
		template.Status = in.Status
	}

	err = l.svcCtx.FreightTemplateModel.Update(l.ctx, template)
	if err != nil {
		l.Logger.Errorf("更新运费模板失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("更新运费模板成功, id=%d", in.Id)
	return &product.Empty{}, nil
}
