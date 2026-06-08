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

type CreateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCategoryLogic {
	return &CreateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCategoryLogic) CreateCategory(in *product.CreateCategoryReq) (*product.Empty, error) {
	if in.Name == "" {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "分类名称不能为空")
	}

	var level int64 = 1
	if in.ParentId > 0 {
		parentCategory, err := l.svcCtx.CategoryModel.FindOne(l.ctx, in.ParentId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "父分类不存在")
			}
			l.Logger.Errorf("查询父分类失败, parentId=%d, err=%v", in.ParentId, err)
			return nil, err
		}
		level = parentCategory.Level + 1
		if level > 3 {
			return nil, errorx.NewBizError(response.ErrCodeParamRange, "分类层级不能超过3级")
		}
	}

	category := &model.Category{
		ParentId: in.ParentId,
		Name:     in.Name,
		Icon:     in.Icon,
		Sort:     in.Sort,
		Level:    level,
		Status:   1,
	}

	_, err := l.svcCtx.CategoryModel.Insert(l.ctx, category)
	if err != nil {
		l.Logger.Errorf("创建分类失败, err=%v", err)
		return nil, err
	}

	l.Logger.Infof("创建分类成功, name=%s, parentId=%d, level=%d", in.Name, in.ParentId, level)
	return &product.Empty{}, nil
}
