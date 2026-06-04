package productadminlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetFreightTemplateLogic) GetFreightTemplate(in *product.IdReq) (*product.FreightTemplate, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "模板ID不能为空")
	}

	template, err := l.svcCtx.FreightTemplateModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeFreightTemplateNotFound, "运费模板不存在")
		}
		l.Logger.Errorf("查询运费模板失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	resp := &product.FreightTemplate{
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
	}

	l.Logger.Infof("获取运费模板成功, id=%d", in.Id)
	return resp, nil
}
