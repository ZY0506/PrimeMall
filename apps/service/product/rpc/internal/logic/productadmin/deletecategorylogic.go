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

type DeleteCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCategoryLogic) DeleteCategory(in *product.IdReq) (*product.Empty, error) {
	if in.Id == 0 {
		return nil, errorx.NewBizError(response.ErrCodeMissingParam, "分类ID不能为空")
	}

	category, err := l.svcCtx.CategoryModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeCategoryNotFound, "分类不存在")
		}
		l.Logger.Errorf("查询分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	children, err := l.svcCtx.CategoryModel.FindAllChildren(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("查询子分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}
	if len(*children) > 0 {
		return nil, errorx.NewBizError(response.ErrCodeCategoryHasChildren, "该分类下存在子分类，无法删除")
	}

	err = l.svcCtx.CategoryModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Logger.Errorf("删除分类失败, id=%d, err=%v", in.Id, err)
		return nil, err
	}

	l.Logger.Infof("删除分类成功, id=%d, name=%s", in.Id, category.Name)
	return &product.Empty{}, nil
}
