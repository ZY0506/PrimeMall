package userlogic

import (
	"context"
	"errors"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

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

// UpdatePassword 修改登录密码（已登录）
func (l *UpdatePasswordLogic) UpdatePassword(in *user.UpdatePasswordReq) (*user.EmptyResp, error) {
	// 查询当前用户
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,userId=%d", userId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	if !pwd.CompareHashAndPassword(u.Password, in.OldPassword) {
		return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "旧密码错误")
	}
	if in.NewPassword != in.ConfirmPassword {
		return nil, errorx.NewBizError(response.ErrCodePasswordNotMatch, "新密码与确认密码不一致")
	}

	// 更新密码
	encodePassword, err := pwd.GenerateFromPassword(in.NewPassword)
	if err != nil {
		l.Logger.Errorf("密码加密失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	u.Password = encodePassword
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		l.Logger.Errorf("更新用户密码失败,userId=%d,err=%v", userId, err)
		return nil, err
	}
	l.Logger.Infof("更新用户密码成功,userId=%d", userId)

	return &user.EmptyResp{}, nil
}
