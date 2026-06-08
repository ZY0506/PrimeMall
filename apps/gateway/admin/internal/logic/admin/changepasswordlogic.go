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

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) error {
	if req.OldPassword == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "旧密码不能为空")
	}
	if req.NewPassword == "" {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "新密码不能为空")
	}
	if len(req.NewPassword) < 6 {
		return errorx.NewBizError(int32(response.ErrCodeInvalidParam), "新密码长度不能少于6位")
	}
	if req.OldPassword == req.NewPassword {
		return errorx.NewBizError(int32(response.ErrCodeSameAsOldPassword), "新密码与旧密码不能一致")
	}
	_, err := l.svcCtx.AdminRpc.UpdatePassword(l.ctx, &admin.UpdatePasswordReq{
		OldPassword: req.OldPassword, NewPassword: req.NewPassword,
	})
	if err != nil {
		l.Logger.Errorf("修改密码失败: %v", err)
		return err
	}
	l.Logger.Infof("修改密码成功")
	return nil
}
