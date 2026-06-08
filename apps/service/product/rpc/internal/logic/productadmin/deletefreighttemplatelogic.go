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

func (l *DeleteFreightTemplateLogic) DeleteFreightTemplate(in *product.IdReq) (*product.Empty, error) {
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

	err = l.svcCtx.FreightTemplateModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("删除运费模板失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("删除运费模板成功, id=%d, name=%s", in.Id, template.Name)
	return &product.Empty{}, nil
}
