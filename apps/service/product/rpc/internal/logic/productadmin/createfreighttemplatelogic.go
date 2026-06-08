package productadminlogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/ZY0506/PrimeMall/common/constants"

	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

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

func (l *CreateFreightTemplateLogic) CreateFreightTemplate(in *product.CreateFreightTemplateReq) (*product.Empty, error) {
	if in.Name == "" {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "模板名称不能为空")
	}
	if in.Type != constants.CALCULATE_FREIGHT_BY_COUNT && in.Type != constants.CALCULATE_FREIGHT_BY_WEIGHT {
		return nil, errorx.NewBizError(response.ErrCodeParamType, "计费方式只能是按件数或按重量")
	}

	if in.IsDefault == 1 {
		defaultTemplate, err := l.svcCtx.FreightTemplateModel.FindDefault(l.ctx)
		if err == nil && defaultTemplate != nil {
			defaultTemplate.IsDefault = 0
			err = l.svcCtx.FreightTemplateModel.Update(l.ctx, defaultTemplate)
			if err != nil {
				l.Logger.Errorf("更新默认模板失败, err=%v", err)
				return nil, err
			}
		}
	}

	template := &model.FreightTemplate{
		Name:                  in.Name,
		Type:                  in.Type,
		DefaultFee:            in.DefaultFee,
		DefaultQuantity:       in.DefaultQuantity,
		ExtraFee:              in.ExtraFee,
		FreeThresholdAmount:   in.FreeThresholdAmount,
		FreeThresholdQuantity: in.FreeThresholdQuantity,
		IsDefault:             in.IsDefault,
		Status:                1,
		DeleteAt:              sql.NullTime{Time: time.Time{}, Valid: false},
	}

	_, err := l.svcCtx.FreightTemplateModel.Insert(l.ctx, template)
	if err != nil {
		l.Logger.Errorf("创建运费模板失败, err=%v", err)
		return nil, err
	}

	l.Logger.Infof("创建运费模板成功, name=%s", in.Name)
	return &product.Empty{}, nil
}
