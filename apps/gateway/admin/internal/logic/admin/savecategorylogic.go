package admin

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/types"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
)

type SaveCategoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveCategoryLogic {
	return &SaveCategoryLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *SaveCategoryLogic) SaveCategory(req *types.SaveCategoryReq) error {
	if req.Name == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "分类名称不能为空")
	}
	if req.Id > 0 {
		_, err := l.svcCtx.AdminProductRpc.UpdateCategory(l.ctx, &admin.AdminUpdateCategoryReq{
			Id: req.Id, Name: req.Name, Icon: req.Icon, Sort: req.Sort,
		})
		if err != nil {
			l.Logger.Errorf("更新分类失败 categoryId=%d: %v", req.Id, err)
			return err
		}
		l.Logger.Infof("更新分类成功 categoryId=%d name=%s", req.Id, req.Name)
		return nil
	}
	_, err := l.svcCtx.AdminProductRpc.CreateCategory(l.ctx, &admin.AdminCreateCategoryReq{
		ParentId: req.ParentId, Name: req.Name, Icon: req.Icon, Sort: req.Sort,
	})
	if err != nil {
		l.Logger.Errorf("创建分类失败: %v", err)
		return err
	}
	l.Logger.Infof("创建分类成功 name=%s", req.Name)
	return nil
}
