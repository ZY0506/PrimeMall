package adminlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(in *admin.UpdatePasswordReq) (*admin.Empty, error) {
	adminId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取操作管理员ID失败,err=%v", err)
		return nil, err
	}

	adminUser, err := l.svcCtx.AdminModel.FindOne(l.ctx, adminId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodeAdminNotFound, "管理员不存在")
		}
		l.Logger.Errorf("查询管理员失败,error=%v", err)
		return nil, err
	}

	if !pwd.CompareHashAndPassword(adminUser.Password, in.OldPassword) {
		return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "旧密码错误")
	}

	hashedPassword, err := pwd.GenerateFromPassword(in.NewPassword)
	if err != nil {
		l.Logger.Errorf("生成密码失败,error=%v", err)
		return nil, err
	}

	adminUser.Password = hashedPassword
	if err := l.svcCtx.AdminModel.Update(l.ctx, adminUser); err != nil {
		l.Logger.Errorf("更新密码失败,error=%v", err)
		return nil, err
	}

	return &admin.Empty{}, nil
}
