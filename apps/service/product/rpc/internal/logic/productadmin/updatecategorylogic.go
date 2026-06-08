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

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCategoryLogic) UpdateCategory(in *product.UpdateCategoryReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "分类ID不能为空")
	}

	category, err := l.svcCtx.CategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "分类不存在")
		}
		l.Logger.Errorf("查询分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	if in.Name != "" {
		category.Name = in.Name
	}
	if in.Icon != "" {
		category.Icon = in.Icon
	}
	if in.Sort != 0 {
		category.Sort = in.Sort
	}
	if in.Status != 0 {
		category.Status = in.Status
	}

	err = l.svcCtx.CategoryModel.Update(l.ctx, category)
	if err != nil {
		l.Logger.Errorf("更新分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("更新分类成功, id=%d", in.Id)
	return &product.Empty{}, nil
}
