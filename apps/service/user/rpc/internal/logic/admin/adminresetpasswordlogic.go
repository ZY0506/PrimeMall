package adminlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminResetPasswordLogic {
	return &AdminResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AdminResetPassword 重置用户密码（管理员强制重置）
func (l *AdminResetPasswordLogic) AdminResetPassword(in *user.AdminResetPasswordReq) (*user.EmptyResp, error) {

	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("用户不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	if userInfo.DeletedAt.Valid {
		l.Logger.Errorf("账号已注销,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "账号已注销")
	}

	// 加密密码
	newPwd, err := pwd.GenerateFromPassword(in.NewPassword)
	if err != nil {
		l.Logger.Errorf("密码加密失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	// 执行更新
	userInfo.Password = newPwd
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Logger.Errorf("更新用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}

	l.Logger.Infof("重置用户密码成功,userId=%d", in.UserId)

	return &user.EmptyResp{}, nil
}
