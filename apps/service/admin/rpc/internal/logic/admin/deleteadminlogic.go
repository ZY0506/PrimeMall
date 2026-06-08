package adminlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAdminLogic {
	return &DeleteAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAdminLogic) DeleteAdmin(in *admin.DeleteAdminReq) (*admin.Empty, error) {
	_, err := l.svcCtx.AdminModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeAdminNotFound, "管理员不存在")
		}
		l.Logger.Errorf("查询管理员失败,error=%v", err)
		return nil, err
	}

	if err := l.svcCtx.AdminModel.Delete(l.ctx, in.Id); err != nil {
		l.Logger.Errorf("删除管理员失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
