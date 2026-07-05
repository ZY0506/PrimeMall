// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package category

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCategoryLogic {
	return &DeleteCategoryLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *DeleteCategoryLogic) DeleteCategory(req *types.IdReq) error {
	if req.Id == 0 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "分类ID不能为空")
	}
	_, err := l.svcCtx.AdminProductRpc.DeleteCategory(l.ctx, &admin.IdReq{Id: req.Id})
	if err != nil {
		l.Logger.Errorf("删除分类失败 categoryId=%d: %v", req.Id, err)
		return err
	}
	l.Logger.Infof("删除分类成功 categoryId=%d", req.Id)
	return nil
}
